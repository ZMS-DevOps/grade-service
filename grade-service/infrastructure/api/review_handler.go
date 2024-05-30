package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/mmmajder/zms-devops-auth-service/application"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/request"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strconv"
)

type ReviewHandler struct {
	reviewService *application.ReviewService
}

func NewReviewHandler(reviewService *application.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

func (handler *ReviewHandler) Init(router *mux.Router) {
	router.HandleFunc(domain.GradeContextPath, handler.AddReview).Methods(http.MethodPost)
	router.HandleFunc(domain.GradeContextPath+"/{id}", handler.UpdateReview).Methods(http.MethodPut)
	router.HandleFunc(domain.GradeContextPath+"/{sub-reviewed}/{type}", handler.GetAllReviewsBySubReviewed).Methods(http.MethodGet)
	router.HandleFunc(domain.GradeContextPath+"/{id}", handler.DeleteReview).Methods(http.MethodDelete)
	router.HandleFunc(domain.GradeContextPath+"/health", handler.GetHealthCheck).Methods(http.MethodGet)
}

func (handler *ReviewHandler) GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, http.StatusOK, domain.HealthCheckMessage)
}

func (handler *ReviewHandler) AddReview(w http.ResponseWriter, r *http.Request) {
	var reviewRequest request.ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&reviewRequest); err != nil {
		handleError(w, http.StatusBadRequest, "Invalid review payload")
		return
	}

	if err := reviewRequest.AreValidRequestData(); err != nil {
		handleError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := handler.reviewService.Add(
		reviewRequest.ReviewType,
		reviewRequest.Comment,
		reviewRequest.Grade,
		reviewRequest.SubReviewer,
		reviewRequest.SubReviewed,
		reviewRequest.ReviewerFullName,
	)

	if err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeResponse(w, http.StatusCreated, response)
}

func (handler *ReviewHandler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	reviewPrimitiveId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	var updateReviewRequest request.UpdateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReviewRequest); err != nil {
		handleError(w, http.StatusBadRequest, "Invalid update review payload")
		return
	}

	if err := updateReviewRequest.AreValidRequestData(); err != nil {
		handleError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = handler.reviewService.Update(
		reviewPrimitiveId,
		updateReviewRequest.ReviewType,
		updateReviewRequest.Comment,
		updateReviewRequest.Grade,
	)

	if err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, nil)
}

func (handler *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	reviewPrimitiveId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	if err := handler.reviewService.Delete(reviewPrimitiveId); err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, nil)
}

func (handler *ReviewHandler) GetAllReviewsBySubReviewed(w http.ResponseWriter, r *http.Request) {
	subReviewed := mux.Vars(r)["sub-reviewed"]
	if subReviewed == "" {
		handleError(w, http.StatusBadRequest, "Invalid ID of reviewed object")
		return
	}

	reviewType, err := strconv.Atoi(mux.Vars(r)["type"])
	if err != nil {
		handleError(w, http.StatusBadRequest, "Invalid number for reviewed type")
		return
	}

	response, err := handler.reviewService.GetAllBySubReviewed(subReviewed, reviewType)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
	}

	writeResponse(w, http.StatusOK, response)
}
