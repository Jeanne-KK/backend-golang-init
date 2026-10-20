package repository

import (
	"context"
	"errors"
	"time"

	"backend/internal/model"

	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	return dbFrom(ctx, r.db).Create(token).Error
}

func (r *RefreshTokenRepository) FindValidByHash(ctx context.Context, hash string, now time.Time) (*model.RefreshToken, error) {
	var token model.RefreshToken
	//	find the valid refresh token by the hash
	err := dbFrom(ctx, r.db).
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hash, now).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *RefreshTokenRepository) RevokeByID(ctx context.Context, id uint, at time.Time) error {
	//	revoke the refresh token by the id
	return dbFrom(ctx, r.db).Model(&model.RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", at).Error
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uint, at time.Time) error {
	//	revoke all refresh tokens for the user
	return dbFrom(ctx, r.db).Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", at).Error
}
