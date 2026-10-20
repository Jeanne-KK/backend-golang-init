package handler

import (
	"net/http"

	"backend/internal/dto"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserService *service.UserService
}

func NewUserHandler(
	userService *service.UserService) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.UserService.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]dto.UserResponse, 0)
	if users != nil {
		for _, u := range *users {
			out = append(out, dto.UserFromModel(&u))
		}
	}
	c.JSON(http.StatusOK, dto.FindUsersResponse{Users: out})
}
