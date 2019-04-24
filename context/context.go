package context

import (
	"context"

	"lenslocked.com/models"
)

const (
	// user a private type to ensure that there's no chance of collision with
	// other context values of the same name (but different type)
	userKey privateKey = "user"
)

type privateKey string

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.User {
	if temp := ctx.Value(userKey); temp != nil {
		if user, ok := temp.(*models.User); ok {
			return user
		}
	}
	return nil
}
