package utils

import (
	"context"
)

type User struct {
	Email string
	Password string
}

type UserStorage interface {
	CreateUser(ctx context.Context, user *User) (string, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
}