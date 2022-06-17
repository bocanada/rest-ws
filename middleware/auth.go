package middleware

import (
	"net/http"
	"strings"

	"github.com/bocanada/rest-ws/helpers"
	"github.com/bocanada/rest-ws/server"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

var (
	NO_AUTH_NEEDED = []string{
		"/",
		"login",
		"signup",
	}
)

func shouldCheckToken(route string) bool {
	for _, p := range NO_AUTH_NEEDED {
		if strings.Contains(route, p) {
			return false
		}
	}
	return true
}

func CheckAuthMiddleware(s server.Server) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !shouldCheckToken(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			_, err := helpers.ParseAppClaims(r.Header.Get("Authorization"), func(t *jwt.Token) (interface{}, error) {
				return []byte(s.Config().JWTSecret), nil
			})
			if err != nil {
				helpers.NewResponseError(err).Send(w, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
