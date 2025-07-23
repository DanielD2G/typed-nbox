package auth

import (
	"context"
	"errors"
	"nbox/internal/application"
	"nbox/internal/domain"
	"net/http"
)

func (a Authn) tryBasicAuth(r *http.Request) (context.Context, error) {
	username, pass, ok := r.BasicAuth()
	if !ok {
		return nil, ErrInvalidAuthHeaderFormat
	}

	user, err := a.repository.ValidatePassword(r.Context(), username, pass)

	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrInvalidPassword) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	return application.NewContextWithUser(r.Context(), application.User{
		Name:  user.Username,
		Roles: user.Roles,
	}), nil

}
