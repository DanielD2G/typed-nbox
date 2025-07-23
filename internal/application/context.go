package application

import "context"

type User struct {
	Name  string   `json:"username"`
	Roles []string `json:"roles"`
}

type userSessionKey struct{}

func NewContextWithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userSessionKey{}, user)
}

func UserFromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userSessionKey{}).(User)
	return u, ok
}
