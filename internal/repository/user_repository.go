package repository

import (
	"context"
	"errors"

	"backend/internal/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	FindAll(ctx context.Context) (*[]model.User, error)
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) FindAll(ctx context.Context) (*[]model.User, error) {
	var users []model.User
	err := dbFrom(ctx, r.db).Where("role = ?", model.RoleAdmin).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return &users, nil
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	return dbFrom(ctx, r.db).Create(user).Error
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	//	find the user by the email
	err := dbFrom(ctx, r.db).Where("email = ?", email).First(&user).Error
	if err != nil {
		//	return nil if the user is not found
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	//	find the user by the id
	err := dbFrom(ctx, r.db).First(&user, id).Error
	if err != nil {
		//	return nil if the user is not found
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	//	count the users by the email
	err := dbFrom(ctx, r.db).Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
