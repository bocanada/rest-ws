package models

import (
	"encoding/json"
	"net/http"
)

type Response[T any] struct {
	Error  string `json:"error,omitempty"`
	Result T      `json:"result,omitempty"`
	Ok     bool   `json:"ok"`
}

func NewResponseError(err error) *Response[any] {
	return &Response[any]{
		Error: err.Error(),
		Ok:    false,
	}
}

func NewResponseOk[T any](res T) *Response[T] {
	return &Response[T]{
		Error:  "",
		Result: res,
		Ok:     true,
	}
}

func (r Response[T]) Send(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(r)
}
