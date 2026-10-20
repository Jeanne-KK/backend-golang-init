package model

import "time"

type RefreshToken struct {
	ID        uint       `json:"id"`
	UserID    uint       `json:"user_id"`
	TokenHash string     `json:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	UserAgent string     `json:"user_agent"`
	IP        string     `json:"ip"`
	CreatedAt time.Time
}
