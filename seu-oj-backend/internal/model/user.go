package model

import "time"

type User struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"column:username" json:"username"`
	UserID       string    `gorm:"column:userid" json:"userid"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
	Role         string    `gorm:"column:role" json:"role"`
	Status       string    `gorm:"column:status" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
