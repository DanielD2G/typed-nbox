package auth

import (
	"context"
	"errors"
	"fmt"
	"nbox/internal/application"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func (a Authn) tryJwt(r *http.Request) (context.Context, error) {

	tokenString, ok := extractAuthValue(r.Header.Get("Authorization"), "Bearer")

	if !ok {
		return nil, ErrInvalidAuthHeaderFormat
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritmo de firma inesperado: %v", token.Header["alg"])
		}
		return a.config.HmacSecretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	var roles []string
	if genericRoles, ok := claims["roles"].([]interface{}); ok {
		for _, role := range genericRoles {
			if roleStr, ok := role.(string); ok {
				roles = append(roles, roleStr)
			}
		}
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, ErrTokenClaimInvalid
	}

	return application.NewContextWithUser(r.Context(), application.User{
		Name:  username,
		Roles: roles,
	}), nil
}
