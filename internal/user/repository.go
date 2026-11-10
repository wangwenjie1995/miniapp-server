package user

import (
	"context"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// ErrUsernameExists 表示用户名与数据库中的记录冲突。
var ErrUsernameExists = errors.New("用户名已存在")

type Repository interface {
	CreateUser(ctx context.Context, username string, passwordHash string) (*User, error)

	FindUserByUsername(ctx context.Context, username string) (*User, error)
}

type MySQLRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) CreateUser(ctx context.Context, username string, passwordHash string) (*User, error) {
	user := &User{
		Username:     username,
		PasswordHash: passwordHash,
	}

	err := gorm.G[User](r.db).Create(ctx, user)

	if err != nil {
		var mysqlErr *mysql.MySQLError

		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 &&
			(strings.HasSuffix(mysqlErr.Message, "for key 'users.username'")) {
			return nil, ErrUsernameExists
		}
		return nil, err
	}
	return user, nil
}

func (r *MySQLRepository) FindUserByUsername(ctx context.Context, username string) (*User, error) {
	user, err := gorm.G[User](r.db).Where("username = ?", username).First(ctx)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
