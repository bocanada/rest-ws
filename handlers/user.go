package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

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

func SignUpHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignUpLoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			models.NewResponseError(err).Send(w, http.StatusBadRequest)
			return
		}
		hashedPasswd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			models.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		id, err := ksuid.NewRandom()
		if err != nil {
			models.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		user := models.User{
			Email:    req.Email,
			Password: string(hashedPasswd),
			ID:       id.String(),
		}
		if err = repository.InsertUser(r.Context(), &user); err != nil {
			models.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		resp := SignUpResponse{Email: user.Email, Id: user.ID}
		models.NewResponseOk(resp).Send(w, http.StatusOK)
	}
}

func LoginHandler(s server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignUpLoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			models.NewResponseError(err).Send(w, http.StatusBadRequest)
			return
		}
		user, err := repository.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			models.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		if user == nil {
			models.NewResponseError(errors.New("invalid credentials")).Send(w, http.StatusUnauthorized)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			models.NewResponseError(errors.New("invalid credentials")).Send(w, http.StatusUnauthorized)
			return
		}
		claims := models.AppClaims{
			UserId: user.ID,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(2 * time.Hour * 24).Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(s.Config().JWTSecret))
		if err != nil {
			models.NewResponseError(err).Send(w, http.StatusInternalServerError)
			return
		}
		resp := LoginResponse{Token: tokenString}
		models.NewResponseOk(resp).Send(w, http.StatusOK)
	}
}

func MeHandler(s server.Server) http.HandlerFunc {
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
			user, err := repository.GetUserById(r.Context(), claims.UserId)
			if err != nil {
				models.NewResponseError(err).Send(w, http.StatusInternalServerError)
				return
			}
			models.NewResponseOk(user).Send(w, http.StatusOK)
		} else {
			models.NewResponseError(err).Send(w, http.StatusInternalServerError)
		}
	}
}
