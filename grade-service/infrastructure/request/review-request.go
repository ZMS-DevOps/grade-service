package request

import (
	"github.com/go-playground/validator/v10"
)

type ReviewRequest struct {
	Comment          string  `json:"comment" validate:"required"`
	Grade            float32 `json:"grade" validate:"required,min=0,max=5"`
	SubReviewer      string  `json:"subReviewer" validate:"required"`
	SubReviewed      string  `json:"subReviewed" validate:"required"`
	ReviewerFullName string  `json:"reviewerFullName"`
	ReviewType       int     `json:"reviewType" validate:"min=0,max=1"`
}

func (request ReviewRequest) AreValidRequestData() error {
	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		return err.(validator.ValidationErrors)
	}

	return nil
}
