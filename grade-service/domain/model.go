package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ReviewType int

const (
	Host ReviewType = iota
	Accommodation
)

type Review struct {
	Id                 primitive.ObjectID `bson:"_id"`
	Comment            string             `bson:"comment"`
	Grade              float32            `bson:"grade"`
	SubReviewer        string             `bson:"sub_reviewer"`
	SubReviewed        string             `bson:"sub_reviewed"`
	ReviewerFullName   string             `bson:"reviewer_full_name"`
	DateOfModification time.Time          `bson:"date_of_modification"`
	Type               ReviewType         `bson:"type"`
}
