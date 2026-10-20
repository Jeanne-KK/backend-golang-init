package model

import "time"

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type User struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `gorm:"type:enum('admin','user');not null;default:user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
