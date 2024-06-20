package application

import (
	"encoding/json"
	"errors"
	booking "github.com/ZMS-DevOps/booking-service/proto"
	"github.com/afiskon/promtail-client/promtail"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mmmajder/zms-devops-grade-service/application/external"
	"github.com/mmmajder/zms-devops-grade-service/domain"
	"github.com/mmmajder/zms-devops-grade-service/infrastructure/dto"
	"github.com/mmmajder/zms-devops-grade-service/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"time"
)

type ReviewService struct {
	store         domain.ReviewStore
	HttpClient    *http.Client
	bookingClient booking.BookingServiceClient
	producer      *kafka.Producer
}

func NewReviewService(store domain.ReviewStore, httpClient *http.Client, producer *kafka.Producer, bookingClient booking.BookingServiceClient, loki promtail.Client) *ReviewService {
	return &ReviewService{
		store:         store,
		HttpClient:    httpClient,
		bookingClient: bookingClient,
		producer:      producer,
	}
}

func (service *ReviewService) Add(reviewType int, comment string, grade float32, reviewerSub string, reviewedSub string, fullNameReviewer string, userId string, span trace.Span, loki promtail.Client) (dto.ReviewDTO, error) {
	if reviewCanCreate := service.userCanReview(reviewType, reviewerSub, reviewedSub, span, loki); reviewCanCreate {
		review := &domain.Review{
			Comment:            comment,
			Grade:              grade,
			SubReviewer:        reviewerSub,
			SubReviewed:        reviewedSub,
			ReviewerFullName:   fullNameReviewer,
			DateOfModification: time.Now(),
			Type:               domain.ReviewType(reviewType),
		}
		util.HttpTraceInfo("Inserting review...", span, loki, "Add", "")
		id, err := service.store.Insert(review)
		if err != nil {
			return dto.ReviewDTO{}, err
		}

		util.HttpTraceInfo("Fetching reviews by sub...", span, loki, "Add", "")
		response, err := service.store.GetAllBySubReviewed(reviewedSub, reviewType)
		log.Printf("new average rating %f", service.getAverageRating(response, span, loki))

		service.produceRatingChanged(reviewType, reviewedSub, service.getAverageRating(response, span, loki), span)
		service.produceNotification(reviewType, reviewedSub, fullNameReviewer, userId, span, loki)

		reviewDTO := dto.FromReview(review)
		reviewDTO.Id = id

		return reviewDTO, nil
	}

	return dto.ReviewDTO{}, errors.New("reviewer doesn't have   already exists")
}

func (service *ReviewService) GetAllBySubReviewed(subReviewed string, reviewType int, span trace.Span, loki promtail.Client) (dto.ReviewReportDTO, error) {
	util.HttpTraceInfo("Fetching reviews by sub...", span, loki, "GetAllBySubReviewed", "")
	response, err := service.store.GetAllBySubReviewed(subReviewed, reviewType)
	if err != nil {
		return dto.ReviewReportDTO{}, err
	}
	averageRating, numberOfStars := service.getReviewReportData(response, span, loki)

	reviewReportDTO := dto.ReviewReportDTO{
		TotalReviews:  len(response),
		AverageRating: averageRating,
		NumberOfStars: numberOfStars,
		Reviews:       *dto.FromReviews(response),
	}

	return reviewReportDTO, nil
}

func (service *ReviewService) Update(id primitive.ObjectID, reviewType int, comment string, grade float32, span trace.Span, loki promtail.Client) error {
	util.HttpTraceInfo("Updating reviews...", span, loki, "Update", "")
	review, err := service.store.Update(id, comment, grade)
	if err != nil {
		return err
	}

	util.HttpTraceInfo("Fetching reviews by sub...", span, loki, "Update", "")
	response, err := service.store.GetAllBySubReviewed(review.SubReviewed, reviewType)
	log.Printf("new average rating %f", service.getAverageRating(response, span, loki))

	service.produceRatingChanged(reviewType, review.SubReviewed, service.getAverageRating(response, span, loki), span)

	return nil
}

func (service *ReviewService) Delete(id primitive.ObjectID, reviewType int, span trace.Span, loki promtail.Client) error {
	util.HttpTraceInfo("Fetching review by id...", span, loki, "Delete", "")
	review, err := service.store.Get(id)
	if err != nil {
		return err
	}
	util.HttpTraceInfo("Deleting review by id...", span, loki, "Delete", "")
	if err = service.store.Delete(id); err != nil {
		return err
	}

	util.HttpTraceInfo("Fetching reviews by sub...", span, loki, "Delete", "")
	response, err := service.store.GetAllBySubReviewed(review.SubReviewed, reviewType)
	service.produceRatingChanged(reviewType, review.SubReviewed, service.getAverageRating(response, span, loki), span)

	return nil
}

func (service *ReviewService) produceRatingChanged(reviewType int, reviewedId string, rating float32, span trace.Span) {
	var topic string
	if reviewType == 0 {
		topic = "host-rating.changed"
	} else {
		topic = "accommodation-rating.changed"
	}

	ratingChangedDTO := dto.RatingChangedDTO{
		Id:     reviewedId,
		Rating: rating,
	}
	message, _ := json.Marshal(ratingChangedDTO)
	err := service.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
	}, nil)

	if err != nil {
		log.Fatalf("Failed to produce message: %s", err)
	}

	service.producer.Flush(4 * 1000)
}

func (service *ReviewService) produceNotification(reviewType int, reviewedId string, reviewerName string, userId string, span trace.Span, loki promtail.Client) {
	var topic string
	var notificationDTO dto.NotificationDTO
	if reviewType == 0 {
		topic = "host-review.created"
		notificationDTO = dto.NotificationDTO{
			UserId:       userId,
			ReviewerName: reviewerName,
		}
	} else {
		topic = "accommodation-review.created"
		notificationDTO = dto.NotificationDTO{
			UserId:          userId,
			AccommodationId: reviewedId,
			ReviewerName:    reviewerName,
		}
	}
	util.HttpTraceInfo("Producing notification for "+topic+"...", span, loki, "produceNotification", "")

	message, _ := json.Marshal(notificationDTO)
	err := service.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
	}, nil)

	if err != nil {
		log.Fatalf("Failed to produce message: %s", err)
	}

	service.producer.Flush(4 * 1000)
}

func (service *ReviewService) getReviewReportData(reviews []*domain.Review, span trace.Span, loki promtail.Client) (float32, []dto.NumberOfStars) {
	util.HttpTraceInfo("Calculating review report data...", span, loki, "getReviewReportData", "")
	var totalGrades float32
	gradeCounts := make([]int, 5)
	if len(reviews) == 0 {
		return 0, dto.GetDefaultNumberOfStars()
	}

	for _, review := range reviews {
		totalGrades += review.Grade
		gradeCounts[ratingToIndex(review.Grade)]++
	}

	averageGrade := totalGrades / float32(len(reviews))

	numberOfStars := []dto.NumberOfStars{
		{"1", gradeCounts[0]},
		{"2", gradeCounts[1]},
		{"3", gradeCounts[2]},
		{"4", gradeCounts[3]},
		{"5", gradeCounts[4]},
	}

	return averageGrade, numberOfStars
}

func (service *ReviewService) getAverageRating(reviews []*domain.Review, span trace.Span, loki promtail.Client) float32 {
	util.HttpTraceInfo("Calculating average rating...", span, loki, "getAverageRating", "")
	var totalGrades float32
	if len(reviews) == 0 {
		return 0
	}

	for _, review := range reviews {
		totalGrades += review.Grade
	}

	return totalGrades / float32(len(reviews))
}

func (service *ReviewService) userCanReview(reviewType int, reviewerSub string, reviewedSub string, span trace.Span, loki promtail.Client) bool {
	var canReview bool
	log.Printf("type %d", reviewType)
	if reviewType == 0 {
		response, err := external.IfGuestCanReviewHost(service.bookingClient, reviewerSub, reviewedSub, span, loki)
		if err != nil {
			return false
		}
		canReview = response.HasReservation
	} else {
		response, err := external.IfGuestCanReviewAccommodation(service.bookingClient, reviewerSub, reviewedSub, span, loki)
		if err != nil {
			return false
		}
		canReview = response.HasReservation
	}

	return canReview
}

func ratingToIndex(rating float32) int {
	switch {
	case rating <= 1:
		return 0
	case rating <= 2:
		return 1
	case rating <= 3:
		return 2
	case rating <= 4:
		return 3
	case rating <= 5:
		return 4
	default:
		return 4
	}
}
