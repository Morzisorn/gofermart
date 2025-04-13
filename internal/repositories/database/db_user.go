package database

import (
	"context"
	"fmt"

	"github.com/morzisorn/gofermart/internal/models"
	database "github.com/morzisorn/gofermart/internal/repositories/database/generated"
)

type UserRepository interface {
	RegisterUser(ctx context.Context, user models.User) error
	GetUser(ctx context.Context, login string) (*models.User, error)
}

type userRepository struct {
	q *database.Queries
}

func NewUserRepository(q *database.Queries) UserRepository {
	return &userRepository{q: q}
}

func (r *userRepository) RegisterUser(ctx context.Context, user models.User) error {
	return r.q.RegisterUser(ctx, database.RegisterUserParams{
		Login:    user.Login,
		Password: user.Password[:],
	})
}

func (r *userRepository) GetUser(ctx context.Context, login string) (*models.User, error) {
	u, err := r.q.GetUser(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("get db user error: %w", err)
	}

	current, err := pgxFloat4ToFloat64(u.Current)
	if err != nil {
		return nil, fmt.Errorf("get db user error: %w", err)
	}

	return &models.User{
		Login:    u.Login,
		Password: [32]byte(u.Password),
		Current:  current,
	}, nil
}
