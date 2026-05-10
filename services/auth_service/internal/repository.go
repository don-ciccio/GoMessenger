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
	AddDeviceToken(ctx context.Context, userID, token string) error
	RemoveDeviceToken(ctx context.Context, userID, token string) error
	UpdateDisplayName(ctx context.Context, userID, displayName string) error
	ListAllUsers(ctx context.Context, excludeID string, limit, offset int) ([]*User, int64, error)
}
