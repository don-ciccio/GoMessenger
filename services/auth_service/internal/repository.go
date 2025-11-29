package auth

import (
	"context"

	authpb "github.com/Miguel-Pezzini/GoMessenger/services/auth_service/internal/pb"
)

type Repository interface {
	Create(ctx context.Context, user *authpb.RegisterRequest) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	SearchByUsername(ctx context.Context, query string, limit int) ([]*User, error)
	GetUsersByIDs(ctx context.Context, ids []string) ([]*User, error)
}
