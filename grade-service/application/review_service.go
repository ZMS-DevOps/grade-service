package application

import (
	"encoding/json"
	booking "github.com/ZMS-DevOps/booking-service/proto"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mmmajder/zms-devops-auth-service/application/external"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/dto"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func NewReviewService(store domain.ReviewStore, httpClient *http.Client, producer *kafka.Producer, bookingClient booking.BookingServiceClient) *ReviewService {
	return &ReviewService{
		store:         store,
		HttpClient:    httpClient,
		bookingClient: bookingClient,
		producer:      producer,
	}
}

func (service *ReviewService) Add(reviewType int, comment string, grade float32, reviewerSub string, reviewedSub string, fullNameReviewer string, userId string) (dto.ReviewDTO, error) {
	log.Printf("before comment %s", comment)
	//if reviewCanCreate := service.userCanReview(reviewType, reviewerSub, reviewedSub); reviewCanCreate {
	review := &domain.Review{
		Comment:            comment,
		Grade:              grade,
		SubReviewer:        reviewerSub,
		SubReviewed:        reviewedSub,
		ReviewerFullName:   fullNameReviewer,
		DateOfModification: time.Now(),
		Type:               domain.ReviewType(reviewType),
	}

	log.Printf("after", comment)
	id, err := service.store.Insert(review)
	if err != nil {
		return dto.ReviewDTO{}, err
	}

	response, err := service.store.GetAllBySubReviewed(reviewedSub, reviewType)
	log.Printf("new average rating %f", service.getAverageRating(response))

	service.produceRatingChanged(reviewType, reviewedSub, service.getAverageRating(response))
	service.produceNotification(reviewType, reviewedSub, fullNameReviewer, userId)

	reviewDTO := dto.FromReview(review)
	reviewDTO.Id = id

	return reviewDTO, nil
	//}

	//return dto.ReviewDTO{}, errors.New("reviewer doesn't have   already exists")

}

func (service *ReviewService) GetAllBySubReviewed(subReviewed string, reviewType int) (dto.ReviewReportDTO, error) {
	response, err := service.store.GetAllBySubReviewed(subReviewed, reviewType)
	if err != nil {
		return dto.ReviewReportDTO{}, err
	}
	averageRating, numberOfStars := service.getReviewReportData(response)

	reviewReportDTO := dto.ReviewReportDTO{
		TotalReviews:  len(response),
		AverageRating: averageRating,
		NumberOfStars: numberOfStars,
		Reviews:       *dto.FromReviews(response),
	}

	return reviewReportDTO, nil
}

func (service *ReviewService) Update(id primitive.ObjectID, reviewType int, comment string, grade float32) error {
	review, err := service.store.Update(id, comment, grade)
	if err != nil {
		return err
	}

	response, err := service.store.GetAllBySubReviewed(review.SubReviewed, reviewType)
	log.Printf("new average rating %f", service.getAverageRating(response))

	service.produceRatingChanged(reviewType, review.SubReviewed, service.getAverageRating(response))

	return nil
}

func (service *ReviewService) Delete(id primitive.ObjectID, reviewType int) error {
	review, err := service.store.Get(id)
	if err != nil {
		return err
	}
	if err = service.store.Delete(id); err != nil {
		return err
	}

	response, err := service.store.GetAllBySubReviewed(review.SubReviewed, reviewType)
	service.produceRatingChanged(reviewType, review.SubReviewed, service.getAverageRating(response))

	return nil
}

func (service *ReviewService) produceRatingChanged(reviewType int, reviewedId string, rating float32) {
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

func (service *ReviewService) produceNotification(reviewType int, reviewedId string, reviewerName string, userId string) {
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

func (service *ReviewService) getReviewReportData(reviews []*domain.Review) (float32, []dto.NumberOfStars) {
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

func (service *ReviewService) getAverageRating(reviews []*domain.Review) float32 {
	var totalGrades float32
	if len(reviews) == 0 {
		return 0
	}

	for _, review := range reviews {
		totalGrades += review.Grade
	}

	return totalGrades / float32(len(reviews))
}

func (service *ReviewService) userCanReview(reviewType int, reviewerSub string, reviewedSub string) bool {
	var canReview bool
	log.Printf("type %d", reviewType)
	if reviewType == 0 {
		response, err := external.IfGuestCanReviewHost(service.bookingClient, reviewerSub, reviewedSub)
		log.Printf("response %d", response.HasReservation)
		if err != nil {
			return false
		}
		canReview = response.HasReservation
	} else {
		response, err := external.IfGuestCanReviewAccommodation(service.bookingClient, reviewerSub, reviewedSub)
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
