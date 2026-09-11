package model

import "time"

type User struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement"`
	Token     string     `gorm:"column:token"`
	Active    bool       `gorm:"column:active;default:true"`
	CreatedAt *time.Time `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string {
	return "users"
}
