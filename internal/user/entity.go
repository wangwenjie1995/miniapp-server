package user

import (
	"time"
)

type User struct {
	ID           uint64 `gorm:"primaryKey"`
	Username     string `gorm:"size:24;not null;uniqueIndex"`
	PasswordHash string `gorm:"size:72;not null" json:"-"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Roles []Role `gorm:"many2many:user_roles;" json:"-"`
}

// 为啥要用users而不是user？
// 因为user是SQL关键字，容易引起歧义。
// 另外，数据库表名一般用复数形式，表示表中有多条记录。
func (User) TableName() string { return "users" }

type Role struct {
	ID          uint64 `gorm:"primaryKey"`
	Code        string
	Name        string
	Description string
	Status      uint8 // 0 禁用，1 启用
	IsBuiltin   uint8 // 0 自定义角色，1 系统内置角色
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Permissions []Permission `gorm:"many2many:role_permissions;"`
	Users       []User       `gorm:"many2many:user_roles;" json:"-"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID          uint64 `gorm:"primaryKey"`
	Code        string
	Name        string
	Module      string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Permission) TableName() string { return "permissions" }

type UserRole struct {
	UserID    uint64 `gorm:"primaryKey"`
	RoleID    uint64 `gorm:"primaryKey"`
	CreatedAt time.Time
}

func (UserRole) TableName() string { return "user_roles" }

type RolePermission struct {
	RoleID       uint64 `gorm:"primaryKey"`
	PermissionID uint64 `gorm:"primaryKey"`
	CreatedAt    time.Time
}

func (RolePermission) TableName() string { return "role_permissions" }
