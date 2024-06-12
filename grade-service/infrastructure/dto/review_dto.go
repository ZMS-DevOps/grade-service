package dto

import (
	"github.com/mmmajder/zms-devops-grade-service/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ReviewDTO struct {
	Id                 primitive.ObjectID `json:"id"`
	Comment            string             `json:"comment"`
	Grade              float32            `json:"grade"`
	SubReviewer        string             `json:"subReviewer"`
	FullName           string             `json:"fullName"`
	DateOfModification time.Time          `json:"dateOfModification"`
}

func FromReviews(reviews []*domain.Review) *[]ReviewDTO {
	reviewDTOs := make([]ReviewDTO, 0, len(reviews))
	for _, review := range reviews {
		dto := FromReview(review)
		reviewDTOs = append(reviewDTOs, dto)
	}
	return &reviewDTOs
}

func FromReview(review *domain.Review) ReviewDTO {
	dto := ReviewDTO{
		Id:                 review.Id,
		Comment:            review.Comment,
		Grade:              review.Grade,
		SubReviewer:        review.SubReviewer,
		FullName:           review.ReviewerFullName,
		DateOfModification: review.DateOfModification,
	}
	return dto
}
