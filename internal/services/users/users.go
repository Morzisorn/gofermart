package users

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/morzisorn/gofermart/internal/hash"
	"github.com/morzisorn/gofermart/internal/logger"
	"github.com/morzisorn/gofermart/internal/models"
	"github.com/morzisorn/gofermart/internal/repositories"
	"go.uber.org/zap"
)

type UserService struct {
	repo repositories.Repository
}

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyRegistered = errors.New("user is already registered")
	ErrIncorrectCredentials  = errors.New("incorrect login or password")
	ErrInternalServerError   = errors.New("internal server error")
)

func NewUserService(repo repositories.Repository) *UserService {
	return &UserService{repo: repo}
}

func (us *UserService) GetUser(ctx context.Context, user *models.User) (*models.User, error) {
	var err error
	user, err = us.repo.GetUser(ctx, user.Login)
	switch err {
	case pgx.ErrNoRows:
		return nil, ErrUserNotFound
	case nil:
		return user, nil
	}
	return nil, err
}

func (us *UserService) RegisterUser(ctx context.Context, user *models.ParseUserRegister) (string, error) {
	_, err := us.GetUser(ctx, &models.User{Login: user.Login})
	switch {
	case err == nil:
		return "", ErrUserAlreadyRegistered
	case !errors.Is(err, ErrUserNotFound):
		return "", err
	}

	hash := hash.GetHash([]byte(user.Password))
	err = us.repo.RegisterUser(ctx, &models.User{
		Login:    user.Login,
		Password: hash,
	})
	if err != nil {
		logger.Log.Error("Sign up user error: ", zap.Error(err))
	}

	token, err := generateToken(user.Login)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (us *UserService) LoginUser(ctx context.Context, user * models.ParseUserRegister) (string, error) {
	dbUser, err := us.GetUser(ctx, &models.User{
		Login: user.Login,
	})
	switch {
	case errors.Is(err, ErrUserNotFound):
		return "", ErrIncorrectCredentials
	case err != nil:
		return "", ErrInternalServerError
	}

	if hash := hash.GetHash([]byte(user.Password)); hash != dbUser.Password {
		return "", ErrIncorrectCredentials
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
