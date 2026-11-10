package user

import "time"

type User struct {
	ID uint64 `gorm:"primaryKey"`
	Username string `gorm:"size:24;not null;uniqueIndex"`
	PasswordHash string `gorm:"size:72;not null" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// 为啥要用users而不是user？
// 因为user是SQL关键字，容易引起歧义。
// 另外，数据库表名一般用复数形式，表示表中有多条记录。
func (User) TableName() string { return "users" }

