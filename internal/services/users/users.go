package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/morzisorn/gofermart/internal/errs"
	"github.com/morzisorn/gofermart/internal/hash"
	"github.com/morzisorn/gofermart/internal/models"
	"github.com/morzisorn/gofermart/internal/repositories"
)

type UserService struct {
	repo repositories.Repository
}

func NewUserService(repo repositories.Repository) *UserService {
	return &UserService{repo: repo}
}

func (us *UserService) GetUser(ctx context.Context, user *models.User) (*models.User, error) {
	var err error
	user, err = us.repo.GetUser(ctx, user.Login)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil, fmt.Errorf("get user error: %w", errs.ErrUserNotFound)
	case err == nil:
		return user, nil
	}
	return nil, fmt.Errorf("get user error: %w", err)
}

func (us *UserService) RegisterUser(ctx context.Context, user *models.ParseUserRegister) (string, error) {
	_, err := us.GetUser(ctx, &models.User{Login: user.Login})
	switch {
	case err == nil:
		return "", fmt.Errorf("register user error: %w", errs.ErrUserAlreadyRegistered)
	case !errors.Is(err, errs.ErrUserNotFound):
		return "", fmt.Errorf("register user error: %w", err)
	}

	hash := hash.GetHash([]byte(user.Password))
	err = us.repo.RegisterUser(ctx, &models.User{
		Login:    user.Login,
		Password: hash,
	})
	if err != nil {
		return "", fmt.Errorf("register user error: %w", err)
	}

	token, err := generateToken(user.Login)
	if err != nil {
		return "", fmt.Errorf("register user error: %w", err)
	}

	return token, nil
}

func (us *UserService) LoginUser(ctx context.Context, user *models.ParseUserRegister) (string, error) {
	dbUser, err := us.GetUser(ctx, &models.User{
		Login: user.Login,
	})
	switch {
	case errors.Is(err, errs.ErrUserNotFound):
		return "", fmt.Errorf("login error: %w", errs.ErrIncorrectCredentials)
	case err != nil:
		return "", fmt.Errorf("login error: %w", errs.ErrInternalServerError)
	}

	if hash := hash.GetHash([]byte(user.Password)); hash != dbUser.Password {
		return "", fmt.Errorf("login error: %w", errs.ErrIncorrectCredentials)
	}

	return generateToken(user.Login)
}

func (us *UserService) GetBalance(ctx context.Context, user *models.User) (*models.UserBalance, error) {
	user, err := us.GetUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &models.UserBalance{
		Current:   user.Current,
		Withdrawn: user.Withdrawn,
	}, nil
}
