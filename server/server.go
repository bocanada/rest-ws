package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Config struct {
	Port        string
	JWTSecret   string
	DatabaseUrl string
}

type Server interface {
	Config() *Config
}

type Broker struct {
	config *Config
	router *mux.Router
}

func (b *Broker) Config() *Config {
	return b.config
}

func NewServer(ctx context.Context, cfg *Config) (*Broker, error) {
	if cfg.Port == "" {
		return nil, errors.New("port is required")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("jwt secret is required")
	}
	if cfg.DatabaseUrl == "" {
		return nil, errors.New("database url is required")
	}

	return &Broker{config: cfg, router: mux.NewRouter()}, nil
}

func (b *Broker) Start(binder func(s Server, r *mux.Router)) {
	b.router = mux.NewRouter()
	binder(b, b.router)
	log.Println("Starting server on port", b.Config().Port)
	if err := http.ListenAndServe(b.config.Port, b.router); err != nil {
		log.Fatal("ListenAndServe:", err.Error())
	}
}
