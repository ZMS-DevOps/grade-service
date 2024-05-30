package persistence

import (
	"context"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DATABASE   = "gradedb"
	COLLECTION = "review"
)

type ReviewMongoDBStore struct {
	reviews *mongo.Collection
}

func NewReviewMongoDBStore(client *mongo.Client) domain.ReviewStore {
	reviews := client.Database(DATABASE).Collection(COLLECTION)
	return &ReviewMongoDBStore{
		reviews: reviews,
	}
}

func (store *ReviewMongoDBStore) Get(id primitive.ObjectID) (*domain.Review, error) {
	filter := bson.M{"_id": id}
	return store.filterOne(filter)
}

func (store *ReviewMongoDBStore) GetAllBySubReviewed(subReviewed string, reviewType int) ([]*domain.Review, error) {
	filter := bson.M{"sub_reviewed": subReviewed, "type": reviewType}

	return store.filter(filter)
}

func (store *ReviewMongoDBStore) Insert(review *domain.Review) (primitive.ObjectID, error) {
	review.Id = primitive.NewObjectID()
	result, err := store.reviews.InsertOne(context.TODO(), review)
	if err != nil {
		return primitive.NilObjectID, err
	}
	review.Id = result.InsertedID.(primitive.ObjectID)
	return review.Id, nil
}

func (store *ReviewMongoDBStore) Delete(id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := store.reviews.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func (store *ReviewMongoDBStore) DeleteAll() {
	store.reviews.DeleteMany(context.TODO(), bson.D{{}})
}

func (store *ReviewMongoDBStore) Update(id primitive.ObjectID, review *domain.Review) error {
	filter := bson.M{"_id": id}
	update := bson.D{
		{"$set", bson.D{
			{"comment", review.Comment},
			{"grade", review.Grade},
			{"date_of_modification", review.DateOfModification},
		}},
	}
	_, err := store.reviews.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (store *ReviewMongoDBStore) filter(filter interface{}) ([]*domain.Review, error) {
	cursor, err := store.reviews.Find(context.TODO(), filter)
	defer cursor.Close(context.TODO())

	if err != nil {
		return nil, err
	}
	return decode(cursor)
}

func (store *ReviewMongoDBStore) filterOne(filter interface{}) (review *domain.Review, err error) {
	result := store.reviews.FindOne(context.TODO(), filter)
	err = result.Decode(&review)
	return
}

func decode(cursor *mongo.Cursor) (reviews []*domain.Review, err error) {
	for cursor.Next(context.TODO()) {
		var review domain.Review
		err = cursor.Decode(&review)
		if err != nil {
			return
		}
		reviews = append(reviews, &review)
	}
	err = cursor.Err()
	return
}
