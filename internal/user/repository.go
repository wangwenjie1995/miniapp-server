package user

import {
	"context"
	"time"

	"gorm.io/gorm"
}

type Repository interface {
	CreateUser(ctx context.Context, username string, passwordHash string) (*User, error)

	FindUserByUsername(ctx context.Context, username string) (*User, error)
}

type MySQLRepository struct{ db *gorm.DB}

func NewRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) CreateUser(ctx context.Context, username string, passwordHash string) (*User, error) {
	user := &User{
		Username: username,
		PasswordHash: passwordHash,
	}

	err := gorm.G[User](r.db).Create(ctx, user)

	if err != nil {
		return nil, err
	}
	return user, nil
}


func