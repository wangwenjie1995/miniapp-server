package user

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, username string, password string) (*User, error) {
	passwordHash := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	return s.repo.CreateUser(ctx, username, passwordHash)
}

func (s *Service) VerifyCredentials(ctx context.Context, username string, password string) (*User, error) {
	user, err := s.repo.FindUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	passwordHash := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	if user.PasswordHash != passwordHash {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}
