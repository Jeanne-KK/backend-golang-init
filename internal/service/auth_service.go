package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/model"
	"backend/internal/pkg/jwt"
	"backend/internal/pkg/password"
	"backend/internal/pkg/tokenhash"
	"backend/internal/repository"
)

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidRefresh     = errors.New("invalid refresh token")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthService struct {
	cfg     *config.Config
	users   repository.UserRepository
	refresh *repository.RefreshTokenRepository
	hasher  *password.Hasher
	jwt     *jwt.Service
	tx      repository.TxManager
}

func NewAuthService(
	cfg *config.Config,
	users repository.UserRepository,
	refresh *repository.RefreshTokenRepository,
	tx repository.TxManager,
	hasher *password.Hasher,
	jwtSvc *jwt.Service,
) *AuthService {
	return &AuthService{
		cfg:     cfg,
		users:   users,
		refresh: refresh,
		hasher:  hasher,
		jwt:     jwtSvc,
		tx:      tx,
	}
}

type AuthTokens struct {
	AccessToken      string
	TokenType        string
	ExpiresIn        int64
	RefreshToken     string
	RefreshExpiresAt time.Time
	User             *model.User
}

type SessionMeta struct {
	UserAgent string
	IP        string
}

func (s *AuthService) Register(ctx context.Context, name, email, password string, meta SessionMeta) (*AuthTokens, error) {
	//	trim the email and convert it to lowercase
	email = strings.ToLower(strings.TrimSpace(email))

	//	check if the email is already taken
	exists, err := s.users.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailTaken
	}

	//	hash the password
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:     strings.TrimSpace(name),
		Email:    email,
		Password: hash,
	}

	//	create the user
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	//	issue the tokens
	return s.issueTokens(ctx, user, meta)
}

func (s *AuthService) Login(ctx context.Context, email, password string, meta SessionMeta) (*AuthTokens, error) {
	//	trim the email and convert it to lowercase
	email = strings.ToLower(strings.TrimSpace(email))

	//	find the user by the email
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	//	verify the password
	ok, err := s.hasher.Verify(password, user.Password)
	if err != nil || !ok {
		return nil, ErrInvalidCredentials
	}

	//	issue the tokens
	return s.issueTokens(ctx, user, meta)
}

func (s *AuthService) Refresh(ctx context.Context, rawRefresh string, meta SessionMeta) (*AuthTokens, error) {
	now := time.Now()
	hash := tokenhash.SHA256Hex(rawRefresh)

	//	find the valid refresh token by the hash
	stored, err := s.refresh.FindValidByHash(ctx, hash, now)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return nil, ErrInvalidRefresh
	}

	//	find the user by the id
	user, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidRefresh
	}

	var tokens *AuthTokens
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.refresh.RevokeByID(txCtx, stored.ID, now); err != nil {
			return err
		}
		tokens, err = s.issueTokens(txCtx, user, meta)
		return err
	})
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uint, rawRefresh string) error {
	now := time.Now()
	if rawRefresh != "" {

		//	hash the refresh token
		hash := tokenhash.SHA256Hex(rawRefresh)

		//	find the valid refresh token by the hash
		stored, err := s.refresh.FindValidByHash(ctx, hash, now)
		if err != nil {
			return err
		}
		if stored != nil && stored.UserID == userID {
			return s.refresh.RevokeByID(ctx, stored.ID, now)
		}
	}
	return s.refresh.RevokeAllForUser(ctx, userID, now)
}

func (s *AuthService) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *AuthService) issueTokens(ctx context.Context, user *model.User, meta SessionMeta) (*AuthTokens, error) {
	//	issue the access token
	access, err := s.jwt.Issue(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	//	generate a new refresh token
	rawRefresh, expiresAt, err := newRefreshToken(s.cfg.JWT.RefreshTTL)
	if err != nil {
		return nil, err
	}

	rt := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenhash.SHA256Hex(rawRefresh),
		ExpiresAt: expiresAt,
		UserAgent: meta.UserAgent,
		IP:        meta.IP,
	}

	if err := s.refresh.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:      access.Token,
		TokenType:        "Bearer",
		ExpiresIn:        access.ExpiresIn,
		RefreshToken:     rawRefresh,
		RefreshExpiresAt: expiresAt,
		User:             user,
	}, nil
}

func newRefreshToken(ttl time.Duration) (string, time.Time, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	return base64.RawURLEncoding.EncodeToString(b), time.Now().Add(ttl), nil
}
