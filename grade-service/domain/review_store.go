package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewStore interface {
	Get(id primitive.ObjectID) (*Review, error)
	GetAllBySubReviewed(subReviewed string, reviewType int) ([]*Review, error)
	Insert(review *Review) (primitive.ObjectID, error)
	Delete(id primitive.ObjectID) error
	DeleteAll()
	Update(id primitive.ObjectID, comment string, grade float32) (*Review, error)
}
