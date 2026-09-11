package repository

import (
	"github.com/sinha-shaurya/agentic-bank/auth/model"
	"gorm.io/gorm"
)

func AddUser(db *gorm.DB, user *model.User) error {
	return db.Create(user).Error
}
