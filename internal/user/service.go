package user

import (
	"context"
	"crypto/md5"
	"fmt"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, username string, password string) (*User, error) {
	passwordHash := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	return s.repo.CreateUser(ctx, username, passwordHash)
}

func (s *Service) FindUserByUsername(ctx context.Context, username string) (*User, error) {
	return s.repo.FindUserByUsername(ctx, username)
}
