package handlers

import (
	"net/http"

	"github.com/bocanada/rest-ws/models"
	"github.com/bocanada/rest-ws/server"
)

type HomeResponse struct {
	Message string `json:"message"`
}

func HomeHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := HomeResponse{Message: "Welcome to my world :)"}
		models.NewResponseOk(resp).Send(w, http.StatusOK)
	}
}
