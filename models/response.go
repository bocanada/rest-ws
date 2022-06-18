package models

import (
	"encoding/json"
	"net/http"
)

type Response[T any] struct {
	Error  string `json:"error,omitempty"`
	Result T      `json:"result,omitempty"`
	Next   string `json:"next,omitempty"`
	Ok     bool   `json:"ok"`
}

func (r Response[T]) Send(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(r)
}
