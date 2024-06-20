package startup

import (
	"fmt"
	booking "github.com/ZMS-DevOps/booking-service/proto"
	"github.com/afiskon/promtail-client/promtail"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gorilla/mux"
	"github.com/mmmajder/zms-devops-grade-service/application"
	"github.com/mmmajder/zms-devops-grade-service/application/external"
	"github.com/mmmajder/zms-devops-grade-service/domain"
	"github.com/mmmajder/zms-devops-grade-service/infrastructure/api"
	"github.com/mmmajder/zms-devops-grade-service/infrastructure/persistence"
	"github.com/mmmajder/zms-devops-grade-service/startup/config"
	"go.mongodb.org/mongo-driver/mongo"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"log"
	"net/http"
)

type Server struct {
	config        *config.Config
	router        *mux.Router
	traceProvider *sdktrace.TracerProvider
	loki          promtail.Client
}

func NewServer(config *config.Config, traceProvider *sdktrace.TracerProvider, loki promtail.Client) *Server {
	return &Server{
		config:        config,
		router:        mux.NewRouter(),
		traceProvider: traceProvider,
		loki:          loki,
	}
}

func (server *Server) Start(producer *kafka.Producer) {
	server.setupHandlers(producer)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", server.config.Port), server.router))
}

func (server *Server) setupHandlers(producer *kafka.Producer) {
	mongoClient := server.initMongoClient()
	reviewStore := server.initReviewStore(mongoClient)
	bookingClient := external.NewBookingClient(server.getBookingAddress())

	reviewService := server.initReviewService(reviewStore, producer, bookingClient)
	reviewHandler := server.initReviewHandler(reviewService)

	reviewHandler.Init(server.router)
}

func (server *Server) initReviewService(store domain.ReviewStore, producer *kafka.Producer, bookingClient booking.BookingServiceClient) *application.ReviewService {

	return application.NewReviewService(store, &http.Client{}, producer, bookingClient, server.loki)
}

func (server *Server) initReviewHandler(authService *application.ReviewService) *api.ReviewHandler {
	return api.NewReviewHandler(authService, server.traceProvider, server.loki)
}

func (server *Server) initMongoClient() *mongo.Client {
	client, err := persistence.GetClient(server.config.DBUsername, server.config.DBPassword, server.config.DBHost, server.config.DBPort)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func (server *Server) initReviewStore(client *mongo.Client) domain.ReviewStore {
	store := persistence.NewReviewMongoDBStore(client)
	store.DeleteAll()
	for _, review := range reviews {
		_, _ = store.Insert(review)
	}
	return store
}

func (server *Server) getBookingAddress() string {
	return fmt.Sprintf("%s:%s", server.config.BookingHost, server.config.BookingPort)
}
