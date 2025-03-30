package database

import (
	"context"

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
		return nil, err
	}

	return &models.User{
		Login:    u.Login,
		Password: [32]byte(u.Password),
		Current:  pgxFloat4ToFloat64(u.Current),
	}, nil
}
