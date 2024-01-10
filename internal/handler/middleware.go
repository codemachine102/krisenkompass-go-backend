package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/config"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"github.com/dgrijalva/jwt-go"
)

func (h *Handler) AuthJWT(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		// Allow unauthenticated users in
		if header == "" {
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			next.ServeHTTP(w, r)
			return
		}

		token, err := jwt.ParseWithClaims(headerParts[1], &domain.Claims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("wrong singing method")
			}
			return []byte(config.Get().App.Key), nil
		})
		if err != nil {
			fmt.Println(err)
			next.ServeHTTP(w, r)
			return
		}

		claims, ok := token.Claims.(*domain.Claims)
		if !ok || !token.Valid {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), domain.UserClaimsKey, claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)

}
