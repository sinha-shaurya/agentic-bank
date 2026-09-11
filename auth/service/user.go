package service

import (
	"context"

	"github.com/sinha-shaurya/agentic-bank/auth/dto"
	"github.com/sinha-shaurya/agentic-bank/auth/model"
	"github.com/sinha-shaurya/agentic-bank/auth/repository"
	"github.com/sinha-shaurya/agentic-bank/internal/obs"
	"github.com/sinha-shaurya/agentic-bank/utils/httperr"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RegisterUser(ctx context.Context, db *gorm.DB) (*dto.UserRegistrationResponse, *httperr.HttpError) {
	// generate token and insert to DB
	generatedToken := GenerateToken()
	user := model.User{Token: Hash(generatedToken)}
	log := obs.Logger(ctx)

	err := repository.AddUser(db, &user)
	if err != nil {
		log.Error("user.register_failed", zap.Error(err))
		apiErr := httperr.NewCode(httperr.ErrInternalServer, err)
		return nil, &apiErr
	}

	// never log the token; only the hash is stored and the raw token is returned once
	log.Info("user.registered", zap.Int64("user_id", user.ID))
	return &dto.UserRegistrationResponse{UserID: user.ID, Token: generatedToken}, nil
}
