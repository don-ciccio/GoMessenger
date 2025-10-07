package auth

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(registerUserRequest RegisterUserRequest) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerUserRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	user := &User{
		Username: registerUserRequest.Username,
		Password: string(hashedPassword),
	}
	if err := s.repo.Create(context.Background(), user); err != nil {
		return "", err
	}
	token, err := createToken(user.Username)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *Service) Authenticate(loginUserRequest LoginUserRequest) (string, error) {
	user, err := s.repo.FindByUsername(context.Background(), loginUserRequest.Username)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginUserRequest.Password))
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}
	return createToken(user.Username)
}
