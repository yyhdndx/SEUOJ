package model

import "time"

type ClassMember struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ClassID   uint64    `gorm:"column:class_id;not null;index" json:"class_id"`
	UserID    uint64    `gorm:"column:user_id;not null;index" json:"user_id"`
	Role      string    `gorm:"column:role;type:enum('student','assistant');default:'student';not null" json:"role"`
	Status    string    `gorm:"column:status;type:enum('active','removed');default:'active';not null" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ClassMember) TableName() string { return "class_members" }
