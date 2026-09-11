package repository

import (
	"github.com/sinha-shaurya/agentic-bank/auth/model"
	"gorm.io/gorm"
)

func AddUser(db *gorm.DB, user *model.User) error {
	return db.Create(user).Error
}

func GetUserByToken(db *gorm.DB, tokenHash string) (*model.User, error) {
	var user model.User
	err := db.Scopes(withTokenString(tokenHash), withActive()).Find(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// scopes
func withTokenString(tokenHash string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("token = ?", tokenHash)
	}
}

func withActive() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("active")
	}
}
