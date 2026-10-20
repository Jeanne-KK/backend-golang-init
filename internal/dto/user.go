package dto

import (
	"time"

	"backend/internal/model"
)

type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type FindUsersResponse struct {
	Users []UserResponse `json:"users"`
}

func UserFromModel(u *model.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}
