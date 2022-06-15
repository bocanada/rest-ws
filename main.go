package main

import (
	"context"
	"log"
	"os"

	"github.com/bocanada/rest-ws/handlers"
	"github.com/bocanada/rest-ws/server"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
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
	r.HandleFunc("/", handlers.HomeHandler(s)).Methods("GET")
}
