package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/bocanada/rest-ws/handlers"
	"github.com/bocanada/rest-ws/middleware"
	"github.com/bocanada/rest-ws/server"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	PORT := os.Getenv("PORT")
	JWT_SECRET := os.Getenv("JWT_SECRET")
	DATABASE_URL := os.Getenv("DATABASE_URL")

	cfg := server.Config{JWTSecret: JWT_SECRET, Port: PORT, DatabaseUrl: DATABASE_URL}
	s, err := server.NewServer(context.Background(), &cfg)
	if err != nil {
		log.Fatal(err)
	}
	s.Start(BindRoutes)
}

func BindRoutes(s server.Server, r *mux.Router) {
	r.Use(middleware.CheckAuthMiddleware(s))
	api := r.PathPrefix("/api/v1").Subrouter()
	r.HandleFunc("/ws", s.Hub().HandleWebSocket)
	r.HandleFunc("/", handlers.HomeHandler(s)).Methods(http.MethodGet)
	r.HandleFunc("/signup", handlers.SignUpHandler(s)).Methods(http.MethodPost)
	r.HandleFunc("/login", handlers.LoginHandler(s)).Methods(http.MethodPost)
	r.HandleFunc("/me", handlers.MeHandler(s)).Methods(http.MethodGet)
	r.HandleFunc("/posts/{id}", handlers.GetPostByIdHandler(s)).Methods(http.MethodGet)
	r.HandleFunc("/posts", handlers.ListPostsHandler(s)).Methods(http.MethodGet)

	api.HandleFunc("/posts", handlers.InsertPostHandler(s)).Methods(http.MethodPost, http.MethodOptions)
	api.HandleFunc("/posts/{id}", handlers.UpdatePostHandler(s)).Methods(http.MethodPatch, http.MethodOptions)
	api.HandleFunc("/posts/{id}", handlers.DeletePostHandler(s)).Methods(http.MethodDelete, http.MethodOptions)
}
