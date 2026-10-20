package router

import (
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/pkg/jwt"
	"backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	AuthHandler    *handler.AuthHandler
	UserHandler    *handler.UserHandler
	UserRepository repository.UserRepository
	JWTService     *jwt.Service
	CORS           gin.HandlerFunc
	ValidateOrigin gin.HandlerFunc
}

func SetupRouter(r *gin.Engine, h *Handler) {
	r.Use(h.CORS)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.AuthHandler.Register)
			auth.POST("/login", h.AuthHandler.Login)
			auth.POST("/refresh", h.ValidateOrigin, h.AuthHandler.Refresh)
			auth.POST("/logout", middleware.OptionalAuth(h.JWTService), h.AuthHandler.Logout)
			auth.GET("/me", middleware.AuthRequired(h.JWTService), h.AuthHandler.Me)
		}

		api.GET("/users",
			middleware.AuthRequired(h.JWTService),
			middleware.RequireRole(h.UserRepository, model.RoleAdmin),
			h.UserHandler.GetUsers,
		)
	}
}
