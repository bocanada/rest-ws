package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/bocanada/rest-ws/helpers"
	"github.com/bocanada/rest-ws/models"
	"github.com/bocanada/rest-ws/repository"
	"github.com/bocanada/rest-ws/server"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/segmentio/ksuid"
)

type UpsertPostRequest struct {
	PostContent string `json:"post_content"`
}

type InsertPostResponse struct {
	Id          string `json:"id"`
	PostContent string `json:"post_content"`
}

var (
	PostNotFound = errors.New("post does not exist")
)

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

		var req UpsertPostRequest
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
		postMessage := models.WebSocketMessage{
			Type:    models.PostCreatedMessage,
			Payload: post,
		}
		s.Hub().Broadcast(postMessage, nil)
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
		if post.Id == "" {
			helpers.NewResponseError(PostNotFound).Send(w, http.StatusNotFound)
			return
		}
		helpers.NewResponseOk(post).Send(w, http.StatusOK)
	}
}

func UpdatePostHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := helpers.ParseAppClaims(r.Header.Get("Authorization"), func(_ *jwt.Token) (interface{}, error) {
			return []byte(s.Config().JWTSecret), nil
		})
		if err != nil {
			var statusCode int
			if errors.Is(err, helpers.InvalidToken) {
				statusCode = http.StatusUnauthorized
			} else {
				statusCode = http.StatusInternalServerError
			}
			helpers.NewResponseError(err).Send(w, statusCode)
			return
		}

		var req UpsertPostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusBadRequest)
			return
		}
		vars := mux.Vars(r)
		post := models.Post{
			Id:          vars["id"],
			PostContent: req.PostContent,
			UserId:      claims.UserId,
		}
		if err = repository.UpdatePost(r.Context(), &post); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helpers.NewResponseError(PostNotFound).Send(w, http.StatusNotFound)
			} else {
				helpers.NewResponseError(err).Send(w, http.StatusInternalServerError)
			}
			return
		}
		helpers.NewResponseOk(InsertPostResponse{Id: post.Id, PostContent: post.PostContent}).Send(w, http.StatusOK)
	}
}

func DeletePostHandler(s server.Server) http.HandlerFunc {
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
		if err = repository.DeletePost(r.Context(), post); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				helpers.NewResponseError(PostNotFound).Send(w, http.StatusNotFound)
			} else {
				helpers.NewResponseError(err).Send(w, http.StatusInternalServerError)
			}
			return
		}
		helpers.NewResponseOk(post).Send(w, http.StatusOK)
	}
}

func stringToInt(value string, def uint64) uint64 {
	v, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return def
	}
	return v
}

func ListPostsHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := helpers.ParseAppClaims(r.Header.Get("Authorization"), func(_ *jwt.Token) (interface{}, error) {
			return []byte(s.Config().JWTSecret), nil
		})
		if err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusUnauthorized)
			return
		}
		params := r.URL.Query()
		after := params.Get("after")
		limit := stringToInt(params.Get("limit"), 100)
		posts, err := repository.ListPosts(r.Context(), limit, after)
		if err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		resp := helpers.NewResponseOk(posts)
		if length := len(posts); length > 1 {
			last := posts[len(posts)-1].Id
			params.Set("after", last)
			r.URL.RawQuery = params.Encode()
			resp.Next = r.URL.String()
		}
		resp.Send(w, http.StatusOK)
	}
}
