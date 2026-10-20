package middleware

import (
	"net/http"

	"backend/internal/dto"
	"backend/internal/repository"

	"github.com/gin-gonic/gin"
)

const ContextUserRoleKey = "userRole"

func RequireRole(users repository.UserRepository, roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	// Create a map of allowed roles
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		// Get the user ID from the context
		raw, ok := c.Get(ContextUserIDKey)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    "unauthorized",
				Message: "unauthorized",
			})
			return
		}
		// Convert the user ID to an integer
		userID, ok := raw.(uint)
		if !ok || userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    "unauthorized",
				Message: "unauthorized",
			})
			return
		}

		// Find the user by ID
		user, err := users.FindByID(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Code:    "internal_error",
				Message: "something went wrong",
			})
			return
		}
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    "unauthorized",
				Message: "unauthorized",
			})
			return
		}

		// Check if the user has the required role
		if _, ok := allowed[user.Role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{
				Code:    "forbidden",
				Message: "insufficient permissions",
			})
			return
		}

		c.Set(ContextUserRoleKey, user.Role)
		c.Next()
	}
}
