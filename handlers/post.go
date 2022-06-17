package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bocanada/rest-ws/helpers"
	"github.com/bocanada/rest-ws/models"
	"github.com/bocanada/rest-ws/repository"
	"github.com/bocanada/rest-ws/server"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/segmentio/ksuid"
)

type InsertPostRequest struct {
	PostContent string `json:"post_content"`
}

type InsertPostResponse struct {
	Id          string `json:"id"`
	PostContent string `json:"post_content"`
}

func InsertPostHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := helpers.ParseAppClaims(r.Header.Get("Authorization"), func(_ *jwt.Token) (interface{}, error) {
			return []byte(s.Config().JWTSecret), nil
		})
		var statusCode int
		if err != nil {
			if errors.Is(err, helpers.InvalidToken) {
				statusCode = http.StatusUnauthorized
			} else {
				statusCode = http.StatusInternalServerError
			}
			helpers.NewResponseError(err).Send(w, statusCode)
			return
		}

		var req InsertPostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusBadRequest)
			return
		}

		id, err := ksuid.NewRandom()
		if err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		post := models.Post{
			Id:          id.String(),
			PostContent: req.PostContent,
			UserId:      claims.UserId,
		}
		if err = repository.InsertPost(r.Context(), &post); err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		helpers.NewResponseOk(InsertPostResponse{Id: post.Id, PostContent: post.PostContent}).Send(w, http.StatusOK)
	}
}

func GetPostByIdHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := helpers.ParseAppClaims(r.Header.Get("Authorization"), func(_ *jwt.Token) (interface{}, error) {
			return []byte(s.Config().JWTSecret), nil
		})
		if err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusUnauthorized)
			return
		}
		vars := mux.Vars(r)
		post, err := repository.GetPostById(r.Context(), vars["id"])
		if err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		helpers.NewResponseOk(post).Send(w, http.StatusOK)
	}
}
