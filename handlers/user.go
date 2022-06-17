package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/bocanada/rest-ws/helpers"
	"github.com/bocanada/rest-ws/models"
	"github.com/bocanada/rest-ws/repository"
	"github.com/bocanada/rest-ws/server"
	"github.com/golang-jwt/jwt"
	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
)

type SignUpLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignUpResponse struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

var (
	InvalidCredentials = errors.New("invalid credentials")
	ExpireTime         = time.Now().Add(2 * time.Hour * 24)
)

func SignUpHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignUpLoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusBadRequest)
			return
		}
		hashedPasswd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		id, err := ksuid.NewRandom()
		if err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		user := models.User{
			Email:    req.Email,
			Password: string(hashedPasswd),
			ID:       id.String(),
		}
		if err = repository.InsertUser(r.Context(), &user); err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		resp := SignUpResponse{Email: user.Email, Id: user.ID}
		helpers.NewResponseOk(resp).Send(w, http.StatusOK)
	}
}

func LoginHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignUpLoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusBadRequest)
			return
		}
		user, err := repository.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		if user == nil {
			helpers.NewResponseError(InvalidCredentials).Send(w, http.StatusUnauthorized)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			helpers.NewResponseError(InvalidCredentials).Send(w, http.StatusUnauthorized)
			return
		}
		claims := helpers.NewAppClaims(user.ID, ExpireTime)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(s.Config().JWTSecret))
		if err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		resp := LoginResponse{Token: tokenString}
		helpers.NewResponseOk(resp).Send(w, http.StatusOK)
	}
}

func MeHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := helpers.ParseAppClaims(r.Header.Get("Authorization"), func(_ *jwt.Token) (interface{}, error) {
			return []byte(s.Config().JWTSecret), nil
		})
		if err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusUnauthorized)
			return
		}
		user, err := repository.GetUserById(r.Context(), claims.UserId)
		if err != nil {
			helpers.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		helpers.NewResponseOk(user).Send(w, http.StatusOK)
	}
}
