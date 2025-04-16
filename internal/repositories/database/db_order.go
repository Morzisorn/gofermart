package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/morzisorn/gofermart/internal/errs"
	"github.com/morzisorn/gofermart/internal/models"
	gen "github.com/morzisorn/gofermart/internal/repositories/database/generated"
)

type OrderRepository interface {
	UploadOrder(ctx context.Context, login, number string) (string, error)
	Withdraw(ctx context.Context, login, number string, sum float64) error
	GetUserOrders(ctx context.Context, login string) (*[]models.Order, error)
	GetUserWithdrawals(ctx context.Context, login string) (*[]models.Withdrawal, error)
	UpdateOrderStatus(ctx context.Context, number, status string) error
	GetOrdersWithStatus(ctx context.Context, status string) (*[]models.Order, error)
	OrderProcessed(ctx context.Context, login, number string, accrual float64) error
	GetUpprocessedOrders(ctx context.Context) (*[]models.Order, error)
	GetOrderByNumber(ctx context.Context, number string) (*models.Order, error)
}

type orderRepository struct {
	q  *gen.Queries
	db *pgxpool.Pool
}

func NewOrderRepository(q *gen.Queries, db *pgxpool.Pool) OrderRepository {
	return &orderRepository{
		q:  q,
		db: db,
	}
}

func (r *orderRepository) UploadOrder(ctx context.Context, login, number string) (string, error) {
	err := r.q.UploadOrder(ctx, gen.UploadOrderParams{
		UserLogin: login,
		Number:    number,
	})
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			order, err := r.GetOrderByNumber(ctx, number)
			if err != nil {
				return "", fmt.Errorf("upload to db order error: %w", err)
			}

			if login == order.UserLogin {
				return "", fmt.Errorf("upload to db order error: %w", errs.ErrOrderAlreadyExist)
			}
			return order.UserLogin, fmt.Errorf("upload to db order error: %w", errs.ErrOrderBelongsAnotherUser)
		}
		return "", fmt.Errorf("upload to db order error: %w", err)
	}
	return login, nil
}

func (r *orderRepository) Withdraw(ctx context.Context, login, number string, sum float64) error {
	user, err := r.q.GetUser(ctx, login)
	if err != nil {
		return err
	}

	if float64(user.Current.Float32) < sum {
		return fmt.Errorf("withdraw error: %w", errs.ErrInsufficientBalance)
	}

	err = withTransaction(ctx, r.db, func(qtx *gen.Queries) error {
		if err := qtx.UploadWithdrawal(ctx, gen.UploadWithdrawalParams{
			Number:    number,
			UserLogin: login,
			Sum:       pgtype.Float4{Float32: float32(sum), Valid: true},
		}); err != nil {
			if strings.Contains(err.Error(), "duplicate key value") {
				return fmt.Errorf("order number is already exist")
			}
			return fmt.Errorf("upload withdrawal error: %w", err)
		}

		return qtx.UpdateUserBalance(ctx, gen.UpdateUserBalanceParams{
			Login:     login,
			Current:   pgtype.Float4{Float32: -float32(sum), Valid: true},
			Withdrawn: pgtype.Float4{Float32: float32(sum), Valid: true},
		})
	})

	if err != nil {
		return fmt.Errorf("withdraw db error: %w", err)
	}
	return nil
}

func (r *orderRepository) GetUserOrders(ctx context.Context, login string) (*[]models.Order, error) {
	dbOrders, err := r.q.GetUserOrders(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("get user orders db error: %w", err)
	}

	return dbToModelOrders(&dbOrders)
}

func (r *orderRepository) GetOrdersWithStatus(ctx context.Context, status string) (*[]models.Order, error) {
	dbOrders, err := r.q.GetOrdersWithStatus(ctx, pgtype.Text{String: status})
	if err != nil {
		return nil, fmt.Errorf("get orders by status error: %w", err)
	}

	return dbToModelOrders(&dbOrders)
}

func (r *orderRepository) GetUpprocessedOrders(ctx context.Context) (*[]models.Order, error) {
	dbOrders, err := r.q.GetUnprocessedOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("get unprocessed orders error: %w", err)
	}

	return dbToModelOrders(&dbOrders)
}

func (r *orderRepository) GetUserWithdrawals(ctx context.Context, login string) (*[]models.Withdrawal, error) {
	dbOrders, err := r.q.GetUserWithdrawals(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("get user withdrawals db error: %w", err)
	}

	return dbToModelWithdrawals(&dbOrders)
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, number, status string) error {
	err := r.q.UpdateOrderStatus(ctx, gen.UpdateOrderStatusParams{
		Number: number,
		Status: pgtype.Text{
			String: status,
			Valid:  true,
		},
	})

	if err != nil {
		return fmt.Errorf("update order status db error: %w", err)
	}
	return nil
}

func (r *orderRepository) OrderProcessed(ctx context.Context, login, number string, accrual float64) error {
	err := withTransaction(ctx, r.db, func(qtx *gen.Queries) error {
		if err := qtx.UpdateOrderAccrual(ctx, gen.UpdateOrderAccrualParams{
			Number: number,
			Accrual: pgtype.Float4{
				Float32: float32(accrual),
				Valid:   true,
			},
		}); err != nil {
			return fmt.Errorf("failed to update accrual. Order number: %s", number)
		}

		if err := qtx.UpdateOrderStatus(ctx, gen.UpdateOrderStatusParams{
			Number: number,
			Status: pgtype.Text{
				String: models.OrderStatusPROCESSED,
				Valid:  true,
			},
		}); err != nil {
			return fmt.Errorf("failed to update order status to PROCESSED. Order number: %s", number)
		}

		if accrual > 0 {
			if err := qtx.UpdateUserBalance(ctx, gen.UpdateUserBalanceParams{
				Login: login,
				Current: pgtype.Float4{
					Float32: float32(accrual),
					Valid:   true,
				},
				Withdrawn: pgtype.Float4{
					Float32: float32(0),
					Valid:   true,
				},
			}); err != nil {
				return fmt.Errorf("failed to update user balance. User login: %s", login)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to finish order prcossing in db: %w", err)
	}

	return nil
}

func withTransaction(ctx context.Context, db *pgxpool.Pool, fn func(q *gen.Queries) error) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	qtx := gen.New(tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) 
		}
	}()

	if err := fn(qtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *orderRepository) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	order, err := r.q.GetOrderByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("get order by number db error: %w", err)
	}

	return dbToModelOrder(&order)
}
