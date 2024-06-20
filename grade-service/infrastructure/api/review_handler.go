package api

import (
	"encoding/json"
	"errors"
	"github.com/afiskon/promtail-client/promtail"
	"github.com/gorilla/mux"
	"github.com/mmmajder/zms-devops-grade-service/application"
	"github.com/mmmajder/zms-devops-grade-service/domain"
	"github.com/mmmajder/zms-devops-grade-service/infrastructure/request"
	"github.com/mmmajder/zms-devops-grade-service/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"net/http"
	"strconv"
)

type ReviewHandler struct {
	reviewService *application.ReviewService
	traceProvider *sdktrace.TracerProvider
	loki          promtail.Client
}

func NewReviewHandler(reviewService *application.ReviewService, traceProvider *sdktrace.TracerProvider, loki promtail.Client) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
		traceProvider: traceProvider,
		loki:          loki,
	}
}

func (handler *ReviewHandler) Init(router *mux.Router) {
	router.HandleFunc(domain.GradeContextPath, handler.AddReview).Methods(http.MethodPost)
	router.HandleFunc(domain.GradeContextPath+"/{id}", handler.UpdateReview).Methods(http.MethodPut)
	router.HandleFunc(domain.GradeContextPath+"/{sub-reviewed}/{type}", handler.GetAllReviewsBySubReviewed).Methods(http.MethodGet)
	router.HandleFunc(domain.GradeContextPath+"/{id}/{type}", handler.DeleteReview).Methods(http.MethodDelete)
	router.HandleFunc(domain.GradeContextPath+"/health", handler.GetHealthCheck).Methods(http.MethodGet)
}

func (handler *ReviewHandler) GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, http.StatusOK, domain.HealthCheckMessage)
}

func (handler *ReviewHandler) AddReview(w http.ResponseWriter, r *http.Request) {
	_, span := handler.traceProvider.Tracer(domain.ServiceName).Start(r.Context(), "add-review-post")
	defer func() { span.End() }()
	var reviewRequest request.ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&reviewRequest); err != nil {
		util.HttpTraceError(err, "invalid review payload", span, handler.loki, "AddReview", "")
		handleError(w, http.StatusBadRequest, "Invalid review payload")
		return
	}

	if err := reviewRequest.AreValidRequestData(); err != nil {
		util.HttpTraceError(err, "invalid request data", span, handler.loki, "AddReview", "")
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
		reviewRequest.HostId,
		span, handler.loki,
	)

	if err != nil {
		util.HttpTraceError(err, "failed to add review", span, handler.loki, "AddReview", "")
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}
	util.HttpTraceInfo("Review added successfully", span, handler.loki, "AddReview", "")

	writeResponse(w, http.StatusCreated, response)
}

func (handler *ReviewHandler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	_, span := handler.traceProvider.Tracer(domain.ServiceName).Start(r.Context(), "update-review-put")
	defer func() { span.End() }()
	id := mux.Vars(r)["id"]
	if id == "" {
		util.HttpTraceError(errors.New("review id can not empty"), "review id can not empty", span, handler.loki, "UpdateReview", "")
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	reviewPrimitiveId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		util.HttpTraceError(err, "invalid review id", span, handler.loki, "UpdateReview", "")
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	var updateReviewRequest request.UpdateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReviewRequest); err != nil {
		util.HttpTraceError(err, "invalid update review payload", span, handler.loki, "UpdateReview", "")
		handleError(w, http.StatusBadRequest, "Invalid update review payload")
		return
	}

	if err := updateReviewRequest.AreValidRequestData(); err != nil {
		util.HttpTraceError(err, "invalid request data", span, handler.loki, "UpdateReview", "")
		handleError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = handler.reviewService.Update(
		reviewPrimitiveId,
		updateReviewRequest.ReviewType,
		updateReviewRequest.Comment,
		updateReviewRequest.Grade,
		span, handler.loki,
	)

	if err != nil {
		util.HttpTraceError(err, "failed to update review", span, handler.loki, "UpdateReview", "")
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.HttpTraceInfo("Review updated successfully", span, handler.loki, "UpdateReview", "")
	writeResponse(w, http.StatusOK, nil)
}

func (handler *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	_, span := handler.traceProvider.Tracer(domain.ServiceName).Start(r.Context(), "delete-review-delete")
	defer func() { span.End() }()
	id := mux.Vars(r)["id"]
	if id == "" {
		util.HttpTraceError(errors.New("review id can not empty"), "review id can not empty", span, handler.loki, "DeleteReview", "")
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	reviewPrimitiveId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		util.HttpTraceError(err, "invalid review id", span, handler.loki, "DeleteReview", "")
		handleError(w, http.StatusBadRequest, domain.InvalidIDErrorMessage)
		return
	}

	reviewType, err := strconv.Atoi(mux.Vars(r)["type"])
	if err != nil {
		util.HttpTraceError(err, "invalid review type", span, handler.loki, "DeleteReview", "")
		handleError(w, http.StatusBadRequest, "Invalid number for reviewed type")
		return
	}

	if err := handler.reviewService.Delete(reviewPrimitiveId, reviewType, span, handler.loki); err != nil {
		util.HttpTraceError(err, "failed to delete review", span, handler.loki, "DeleteReview", "")
		handleError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.HttpTraceInfo("Unavailability period added successfully", span, handler.loki, "DeleteReview", "")
	writeResponse(w, http.StatusOK, nil)
}

func (handler *ReviewHandler) GetAllReviewsBySubReviewed(w http.ResponseWriter, r *http.Request) {
	_, span := handler.traceProvider.Tracer(domain.ServiceName).Start(r.Context(), "get-all-reviews-by-sub-reviewed-get")
	defer func() { span.End() }()
	subReviewed := mux.Vars(r)["sub-reviewed"]
	if subReviewed == "" {
		util.HttpTraceError(errors.New("review id can not empty"), "review id can not empty", span, handler.loki, "GetAllReviewsBySubReviewed", "")
		handleError(w, http.StatusBadRequest, "Invalid ID of reviewed object")
		return
	}

	reviewType, err := strconv.Atoi(mux.Vars(r)["type"])
	if err != nil {
		util.HttpTraceError(err, "invalid review type", span, handler.loki, "GetAllReviewsBySubReviewed", "")
		handleError(w, http.StatusBadRequest, "Invalid number for reviewed type")
		return
	}

	response, err := handler.reviewService.GetAllBySubReviewed(subReviewed, reviewType, span, handler.loki)
	if err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
	}

	util.HttpTraceInfo("Successfully fetched all reviews by sub", span, handler.loki, "GetAllReviewsBySubReviewed", "")
	writeResponse(w, http.StatusOK, response)
}
