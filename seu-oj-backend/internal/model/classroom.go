package model

import "time"

type Classroom struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:name;size:255;not null" json:"name"`
	Description string    `gorm:"column:description;type:mediumtext" json:"description"`
	JoinCode    string    `gorm:"column:join_code;size:32;uniqueIndex;not null" json:"join_code"`
	TeacherID   uint64    `gorm:"column:teacher_id;not null;index" json:"teacher_id"`
	Status      string    `gorm:"column:status;type:enum('active','archived');default:'active';not null" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Classroom) TableName() string { return "classes" }
