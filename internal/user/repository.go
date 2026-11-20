package user

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	ListUsers(ctx context.Context) ([]UserResponse, error)
	CreateUser(ctx context.Context, username string, passwordHash string) (*User, error)

	FindUserByUsername(ctx context.Context, username string) (*User, error)

	FindUserByID(ctx context.Context, id uint64) (*User, error)
	FindActiveRoles(ctx context.Context, userID uint64) ([]Role, error)
	FindUserPermission(ctx context.Context, userID uint64) ([]RolePermissionInfo, error)

	HasPermission(ctx context.Context, userID uint64, permissionCode string) (bool, error)
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

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1.创建用户；成功后，GORM会把生成的ID写入user.ID
		err := gorm.G[User](tx).Create(ctx, user)

		if err != nil {
			// 当前 users 表的唯一业务字段是 username。
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrUsernameExists
			}
			return err
		}

		//	2.查找普通用户角色
		role, err := gorm.G[Role](tx).Where("code = ?", "user").First(ctx)
		if err != nil {
			return err
		}

		// 3.创建用户和角色的关联
		userRole := &UserRole{
			UserID: user.ID,
			RoleID: role.ID,
		}

		err = gorm.G[UserRole](tx).Create(ctx, userRole)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
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

func (r *MySQLRepository) FindUserByID(ctx context.Context, id uint64) (*User, error) {
	user, err := gorm.G[User](r.db).Where("id = ?", id).First(ctx)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *MySQLRepository) FindUserRoles(ctx context.Context, userID uint64) ([]UserRole, error) {
	userRoles, err := gorm.G[UserRole](r.db).Where("user_id = ?", userID).Find(ctx)
	if err != nil {
		return nil, err
	}

	return userRoles, nil
}

// FindActiveRoles 查询用户已启用的角色，包括没有权限的角色。
func (r *MySQLRepository) FindActiveRoles(ctx context.Context, userID uint64) ([]Role, error) {
	var roles []Role
	err := gorm.G[any](r.db).Raw(`
		select roles.* from roles
		join user_roles on user_roles.role_id = roles.id
		where user_roles.user_id = ? and roles.status = 1
		order by roles.code
	`, userID).Scan(ctx, &roles)
	return roles, err
}

func (r *MySQLRepository) FindRolePermissions(ctx context.Context, roleID uint64) ([]RolePermission, error) {
	rolePermissions, err := gorm.G[RolePermission](r.db).Where("role_id = ?", roleID).Find(ctx)
	if err != nil {
		return nil, err
	}

	return rolePermissions, nil
}

func (r *MySQLRepository) HasPermission(ctx context.Context, userID uint64, permissionCode string) (bool, error) {
	var allowed bool
	err := gorm.G[any](r.db).Raw(`
		select exists (
			select 1 from user_roles
			join roles on roles.id = user_roles.role_id
			join role_permissions on role_permissions.role_id = roles.id
			join permissions on permissions.id = role_permissions.permission_id
			where user_roles.user_id = ? and roles.status = 1 and permissions.code = ?
		)
	`, userID, permissionCode).Scan(ctx, &allowed)
	return allowed, err
}

func (r *MySQLRepository) FindUserPermission(ctx context.Context, userID uint64) ([]RolePermissionInfo, error) {
	var rolePermissions []RolePermissionInfo

	err := gorm.G[any](r.db).Raw(`
		select distinct
			roles.id as role_id,
			permissions.id as permission_id,
			permissions.code
		from user_roles
		join roles on roles.id = user_roles.role_id
		join role_permissions on role_permissions.role_id = roles.id
		join permissions on permissions.id = role_permissions.permission_id
		where user_roles.user_id = ?
			and roles.status = 1
	`, userID).Scan(ctx, &rolePermissions)
	if err != nil {
		return nil, err
	}

	return rolePermissions, nil
}

func (r *MySQLRepository) FindUsersByRole(ctx context.Context, roleID uint64) ([]UserResponse, error) {
	// 1.Preload写法
	// 需要再entity.go中定义 Users       []User       `gorm:"many2many:user_roles;" json:"-"`
	// role, err := gorm.G[Role](r.db).Preload("Users", nil).Where("id = ?", roleID).First(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// users := make([]UserResponse, 0, len(role.Users))
	// for _, user := range role.Users {
	// 	users = append(users, UserResponse{
	// 		ID: user.ID,
	// 		Username: user.Username,
	// 	})
	// }
	// return users, nil

	// 2.sql写法
	var users []UserResponse
	err := gorm.G[any](r.db).Raw(`
		select
			users.id,
			users.username
		from users
		join user_roles on  user_roles.user_id = users.id
		where user_roles.role_id = ?
	`, roleID).Scan(ctx, &users)

	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *MySQLRepository) ListUsers(ctx context.Context) ([]UserResponse, error) {
	return gorm.G[UserResponse](r.db).Table("users").Select("id, username").Order("id").Find(ctx)
}
