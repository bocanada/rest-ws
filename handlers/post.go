package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bocanada/rest-ws/models"
	"github.com/bocanada/rest-ws/repository"
	"github.com/bocanada/rest-ws/server"
	"github.com/golang-jwt/jwt"
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
		tokenString := strings.TrimSpace(r.Header.Get("Authorization"))
		token, err := jwt.ParseWithClaims(tokenString, &models.AppClaims{}, func(_ *jwt.Token) (interface{}, error) {
			return []byte(s.Config().JWTSecret), nil
		})
		if err != nil {
			models.NewResponseError(err).Send(w, http.StatusUnauthorized)
			return
		}
		if claims, ok := token.Claims.(*models.AppClaims); ok && token.Valid {
			var req InsertPostRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				models.NewResponseError(err).Send(w, http.StatusBadRequest)
				return
			}
			id, err := ksuid.NewRandom()
			if err != nil {
				models.NewResponseError(err).Send(w, http.StatusInternalServerError)
				return
			}
			post := models.Post{
				Id:          id.String(),
				PostContent: req.PostContent,
				UserId:      claims.UserId,
			}
			if err = repository.InsertPost(r.Context(), &post); err != nil {
				models.NewResponseError(err).Send(w, http.StatusInternalServerError)
				return
			}
			models.NewResponseOk(InsertPostResponse{Id: post.Id, PostContent: post.PostContent}).Send(w, http.StatusOK)
		} else {
			models.NewResponseError(err).Send(w, http.StatusInternalServerError)
		}
	}
}
