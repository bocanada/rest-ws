package helpers

import (
	"errors"
	"strings"
	"time"

	"github.com/bocanada/rest-ws/models"
	"github.com/golang-jwt/jwt"
)

var (
	InvalidToken = errors.New("invalid token")
)

func NewAppClaims(userId string, expiresAt time.Time) *models.AppClaims {
	return &models.AppClaims{UserId: userId, StandardClaims: jwt.StandardClaims{
		ExpiresAt: expiresAt.Unix(),
	}}
}

func ParseAppClaims(tokenString string, keyFunc func(*jwt.Token) (interface{}, error)) (*models.AppClaims, error) {
	tokenString = strings.TrimSpace(tokenString)
	token, err := jwt.ParseWithClaims(tokenString, &models.AppClaims{}, keyFunc)
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*models.AppClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, InvalidToken
	}
}

func NewResponseError(err error) *models.Response[any] {
	return &models.Response[any]{
		Error: err.Error(),
		Ok:    false,
	}
}

func NewResponseOk[T any](res T) *models.Response[T] {
	return &models.Response[T]{
		Error:  "",
		Result: res,
		Ok:     true,
	}
}
