package service

import (
	"context"

	"github.com/sinha-shaurya/agentic-bank/auth/dto"
	"github.com/sinha-shaurya/agentic-bank/utils/httperr"
	"gorm.io/gorm"
)

func RegisterUser(ctx context.Context, db *gorm.DB) (*dto.UserRegistrationResponse, *httperr.HttpError) {
	// method skeleton
	return nil, nil
}
