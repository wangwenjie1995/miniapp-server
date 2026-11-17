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

func (s *Service) GetUserByID(ctx context.Context, id uint64) (*User, error) {
	user, err := s.repo.FindUserByID(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetRoleAndPermissionCodes 返回启用角色和去重后的权限编码。
func (s *Service) GetRoleAndPermissionCodes(ctx context.Context, userID uint64) ([]string, []string, error) {
	roles, err := s.repo.FindActiveRoles(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	entries, err := s.repo.FindUserPermission(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	roleCodes := make([]string, 0, len(roles))
	for _, role := range roles {
		roleCodes = append(roleCodes, role.Code)
	}
	permissionCodes := make([]string, 0, len(entries))
	seen := make(map[string]bool)
	for _, entry := range entries {
		if !seen[entry.PermissionCode] {
			seen[entry.PermissionCode] = true
			permissionCodes = append(permissionCodes, entry.PermissionCode)
		}
	}
	return roleCodes, permissionCodes, nil
}
