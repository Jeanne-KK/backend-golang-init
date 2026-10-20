package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type AccessToken struct {
	Token     string
	ExpiresIn int64
}

type Service struct {
	secret []byte
	ttl    time.Duration
}

func NewService(cfg config.JWTConfig) *Service {
	return &Service{
		secret: []byte(cfg.AccessSecret),
		ttl:    cfg.AccessTTL,
	}
}

func (s *Service) Issue(userID uint, email string) (*AccessToken, error) {
	now := time.Now()
	exp := now.Add(s.ttl)
	claims := AccessClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			ID:        randomJTI(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	//	create a new jwt token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return nil, err
	}
	return &AccessToken{
		Token:     signed,
		ExpiresIn: int64(s.ttl.Seconds()),
	}, nil
}

func (s *Service) Parse(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		//	check if the signing method is HS256
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*AccessClaims)
	//	check if the claims are valid
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func UserIDFromClaims(claims *AccessClaims) (uint, error) {
	//	Get the user ID from the claims
	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return 0, errors.New("invalid subject")
	}
	return uint(id), nil
}

func randomJTI() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
