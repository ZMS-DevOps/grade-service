package request

import (
	"github.com/go-playground/validator/v10"
)

type UpdateReviewRequest struct {
	Comment    string  `json:"comment" validate:"required"`
	Grade      float32 `json:"grade" validate:"required,min=0,max=5"`
	ReviewType int     `json:"reviewType" validate:"min=0,max=1"`
}

func (request UpdateReviewRequest) AreValidRequestData() error {
	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		return err.(validator.ValidationErrors)
	}

	return nil
}
