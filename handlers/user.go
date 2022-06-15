package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bocanada/rest-ws/models"
	"github.com/bocanada/rest-ws/repository"
	"github.com/bocanada/rest-ws/server"
	"github.com/segmentio/ksuid"
)

type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignUpResponse struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

func SignUpHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignUpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := ksuid.NewRandom()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user := models.User{
			Email:    req.Email,
			Password: req.Password,
			ID:       id.String(),
		}
		if err = repository.InsertUser(r.Context(), &user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SignUpResponse{Email: user.Email, Id: user.ID})
	}
}
