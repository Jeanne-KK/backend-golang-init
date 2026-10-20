package handler

import (
	"errors"
	"net/http"
	"strings"

	"backend/internal/config"
	"backend/internal/dto"
	"backend/internal/middleware"
	authcookie "backend/internal/pkg/cookie"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	AuthService *service.AuthService
	Config      *config.Config
}

func NewAuthHandler(authSvc *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		AuthService: authSvc,
		Config:      cfg,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !bindJSON(c, &req) {
		return
	}

	tokens, err := h.AuthService.Register(c.Request.Context(), req.Name, req.Email, req.Password, sessionMeta(c))
	if err != nil {
		writeAuthError(c, err)
		return
	}
	writeAuthSuccess(c, h.Config, tokens)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !bindJSON(c, &req) {
		return
	}

	tokens, err := h.AuthService.Login(c.Request.Context(), req.Email, req.Password, sessionMeta(c))
	if err != nil {
		writeAuthError(c, err)
		return
	}
	writeAuthSuccess(c, h.Config, tokens)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	raw, err := authcookie.RefreshTokenFromRequest(c, h.Config.Cookie)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "unauthorized",
			Message: "missing refresh token",
		})
		return
	}

	tokens, err := h.AuthService.Refresh(c.Request.Context(), raw, sessionMeta(c))
	if err != nil {
		writeAuthError(c, err)
		return
	}
	writeAuthSuccess(c, h.Config, tokens)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var userID uint
	if id, ok := c.Get(middleware.ContextUserIDKey); ok {
		if uid, ok := id.(uint); ok {
			userID = uid
		}
	}

	raw, _ := authcookie.RefreshTokenFromRequest(c, h.Config.Cookie)
	if userID != 0 {
		_ = h.AuthService.Logout(c.Request.Context(), userID, raw)
	}

	authcookie.ClearRefreshToken(c, h.Config.Cookie)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "unauthorized",
			Message: "unauthorized",
		})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "unauthorized",
			Message: "unauthorized",
		})
		return
	}

	user, err := h.AuthService.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.UserFromModel(user))
}

func writeAuthSuccess(c *gin.Context, cfg *config.Config, tokens *service.AuthTokens) {
	// Set the refresh token in the cookie
	authcookie.SetRefreshToken(c, cfg.Cookie, tokens.RefreshToken, tokens.RefreshExpiresAt)

	// Set the access token in the response
	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken: tokens.AccessToken,
		TokenType:   tokens.TokenType,
		ExpiresIn:   tokens.ExpiresIn,
		User:        dto.UserFromModel(tokens.User),
	})
}

func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrEmailTaken):
		c.JSON(http.StatusConflict, dto.ErrorResponse{
			Code:    "email_taken",
			Message: "email already registered",
		})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "invalid_credentials",
			Message: "invalid email or password",
		})
	case errors.Is(err, service.ErrInvalidRefresh):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "invalid_refresh",
			Message: "invalid or expired refresh token",
		})
	case errors.Is(err, service.ErrUserNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:    "not_found",
			Message: "user not found",
		})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "internal_error",
			Message: "something went wrong",
		})
	}
}

func sessionMeta(c *gin.Context) service.SessionMeta {
	return service.SessionMeta{
		UserAgent: c.GetHeader("User-Agent"),
		IP:        c.ClientIP(),
	}
}

func bindJSON(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		writeBindError(c, err)
		return false
	}
	return true
}

func writeBindError(c *gin.Context, err error) {
	details := map[string]string{}
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		for _, fe := range verrs {
			field := strings.ToLower(fe.Field())
			details[field] = validationMessage(fe)
		}
	}
	c.JSON(http.StatusBadRequest, dto.ErrorResponse{
		Code:    "validation_error",
		Message: "invalid request body",
		Details: details,
	})
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	default:
		return "is invalid"
	}
}
