package repositories

import (
	"context"

	"github.com/morzisorn/gofermart/internal/models"
	"github.com/morzisorn/gofermart/internal/repositories/database"
)

type Repository interface {
	RegisterUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, login string) (*models.User, error)

	UploadOrder(ctx context.Context, login, number string) (string, error)
	UpdateOrderStatus(ctx context.Context, number, status string) error
	Withdraw(ctx context.Context, login, number string, sum float64) error
	GetUserOrders(ctx context.Context, login string) (*[]models.Order, error)
	GetUserWithdrawals(ctx context.Context, login string) (*[]models.Withdrawal, error)
	GetOrdersWithStatus(ctx context.Context, status string) (*[]models.Order, error)
	OrderProcessed(ctx context.Context, login, number string, accrual float64) error
	GetUpprocessedOrders(ctx context.Context) (*[]models.Order, error)
	GetOrderByNumber(ctx context.Context, number string) (*models.Order, error)
}

type DBRepository struct {
	users  database.UserRepository
	orders database.OrderRepository
}

func (r *DBRepository) RegisterUser(ctx context.Context, user *models.User) error {
	return r.users.RegisterUser(ctx, *user)
}

func (r *DBRepository) GetUser(ctx context.Context, login string) (*models.User, error) {
	return r.users.GetUser(ctx, login)
}

func (r *DBRepository) UploadOrder(ctx context.Context, login, number string) (string, error) {
	return r.orders.UploadOrder(ctx, login, number)
}

func (r *DBRepository) UpdateOrderStatus(ctx context.Context, number, status string) error {
	return r.orders.UpdateOrderStatus(ctx, number, status)
}

func (r *DBRepository) GetOrdersWithStatus(ctx context.Context, status string) (*[]models.Order, error) {
	return r.orders.GetOrdersWithStatus(ctx, status)
}

func (r *DBRepository) Withdraw(ctx context.Context, login, number string, sum float64) error {
	return r.orders.Withdraw(ctx, login, number, sum)
}

func (r *DBRepository) GetUserOrders(ctx context.Context, login string) (*[]models.Order, error) {
	return r.orders.GetUserOrders(ctx, login)
}

func (r *DBRepository) GetUserWithdrawals(ctx context.Context, login string) (*[]models.Withdrawal, error) {
	return r.orders.GetUserWithdrawals(ctx, login)
}

func (r *DBRepository) OrderProcessed(ctx context.Context, login, number string, accrual float64) error {
	return r.orders.OrderProcessed(ctx, login, number, accrual)
}
func (r *DBRepository) GetUpprocessedOrders(ctx context.Context) (*[]models.Order, error) {
	return r.orders.GetUpprocessedOrders(ctx)
}

func (r *DBRepository) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	return r.orders.GetOrderByNumber(ctx, number)
}