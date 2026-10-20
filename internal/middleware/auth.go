package middleware

import (
	"net/http"
	"strings"

	"backend/internal/dto"
	"backend/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey    = "userID"
	ContextUserEmailKey = "userEmail"
)

func AuthRequired(jwtSvc *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the authorization header
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    "unauthorized",
				Message: "missing authorization header",
			})
			return
		}

		// Split the header into parts
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    "unauthorized",
				Message: "invalid authorization header",
			})
			return
		}

		// Parse the token
		claims, err := jwtSvc.Parse(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    "unauthorized",
				Message: "invalid or expired access token",
			})
			return
		}

		// Get the user ID from the claims
		userID, err := jwt.UserIDFromClaims(claims)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    "unauthorized",
				Message: "invalid access token",
			})
			return
		}

		// Set the user ID and email in the context
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextUserEmailKey, claims.Email)
		c.Next()
	}
}

func OptionalAuth(jwtSvc *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.Next()
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Next()
			return
		}
		claims, err := jwtSvc.Parse(parts[1])
		if err == nil {
			userID, uidErr := jwt.UserIDFromClaims(claims)
			if uidErr == nil {
				c.Set(ContextUserIDKey, userID)
				c.Set(ContextUserEmailKey, claims.Email)
			}
		}
		c.Next()
	}
}
