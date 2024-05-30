package startup

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mmmajder/zms-devops-auth-service/application"
	"github.com/mmmajder/zms-devops-auth-service/domain"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/api"
	"github.com/mmmajder/zms-devops-auth-service/infrastructure/persistence"
	"github.com/mmmajder/zms-devops-auth-service/startup/config"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
)

type Server struct {
	config *config.Config
	router *mux.Router
}

func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
		router: mux.NewRouter(),
	}
}

func (server *Server) Start() {
	server.setupHandlers()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", server.config.Port), server.router))
}

func (server *Server) setupHandlers() {
	mongoClient := server.initMongoClient()
	reviewStore := server.initReviewStore(mongoClient)
	authService := server.initReviewService(reviewStore)
	authHandler := server.initReviewHandler(authService)
	authHandler.Init(server.router)
}

func (server *Server) initReviewService(store domain.ReviewStore) *application.ReviewService {

	return application.NewReviewService(store, &http.Client{})
}

func (server *Server) initReviewHandler(authService *application.ReviewService) *api.ReviewHandler {
	return api.NewReviewHandler(authService)
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
