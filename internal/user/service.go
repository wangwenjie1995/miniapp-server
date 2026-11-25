package user

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func createUser(ctx context.Context, repo Repository, username string, password string) (*User, error) {
	passwordHash := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	return repo.CreateUser(ctx, username, passwordHash)
}

func verifyCredentials(ctx context.Context, repo Repository, username string, password string) (*User, error) {
	user, err := repo.FindUserByUsername(ctx, username)
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

func getUserByID(ctx context.Context, repo Repository, id uint64) (*User, error) {
	user, err := repo.FindUserByID(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetRoleAndPermissionCodes 返回启用角色和去重后的权限编码。
func GetRoleAndPermissionCodes(ctx context.Context, repo Repository, userID uint64) ([]string, []string, error) {
	roles, err := repo.FindActiveRoles(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	entries, err := repo.FindUserPermission(ctx, userID)
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
		if !seen[entry.Code] {
			seen[entry.Code] = true
			permissionCodes = append(permissionCodes, entry.Code)
		}
	}
	return roleCodes, permissionCodes, nil
}
