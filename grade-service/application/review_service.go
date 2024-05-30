package application

import (
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/dto"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"time"
)

type ReviewService struct {
	store      domain.ReviewStore
	HttpClient *http.Client
}

func NewReviewService(store domain.ReviewStore, httpClient *http.Client) *ReviewService {
	return &ReviewService{
		store:      store,
		HttpClient: httpClient,
	}
}

func (service *ReviewService) Add(reviewType int, comment string, grade float32, reviewerSub string, reviewedSub string, fullNameReviewer string) (dto.ReviewDTO, error) {
	//todo: call booking-service to check if reviewerSub has DONE reservation for reviewedSub
	review := &domain.Review{
		Comment:            comment,
		Grade:              grade,
		SubReviewer:        reviewerSub,
		SubReviewed:        reviewedSub,
		ReviewerFullName:   fullNameReviewer,
		DateOfModification: time.Now(),
	}

	id, err := service.store.Insert(review)
	if err != nil {
		return dto.ReviewDTO{}, err
	}

	service.updateAverageRatingInServices(reviewType, grade)

	reviewDTO := dto.FromReview(review)
	reviewDTO.Id = id
	return reviewDTO, nil
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

func (service *ReviewService) updateAverageRatingInServices(reviewType int, grade float32) {
	switch reviewType {
	case 0:
		//todo: call user-service to update averageRating for reviewerSub
	case 1:
		//todo: call accomodation-service to update averageRating for reviewerSub
	}
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
		log.Printf("total grades: %.2f\n", totalGrades)
		log.Printf("rating: %d\n", ratingToIndex(review.Grade))
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

func (service *ReviewService) Update(id primitive.ObjectID, reviewType int, comment string, grade float32) error {
	review, err := service.store.Get(id)
	review.Comment = comment
	review.Grade = grade
	review.DateOfModification = time.Now()
	err = service.store.Update(id, review)
	if err != nil {
		return err
	}

	service.updateAverageRatingInServices(reviewType, grade)

	return nil
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
