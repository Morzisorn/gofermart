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
	"github.com/morzisorn/gofermart/internal/logger"
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
				return "", err
			}

			if login == order.UserLogin {
				return "", errs.ErrOrderAlreadyExist
			}
			return order.UserLogin, errs.ErrOrderBelongsAnotherUser
		}
		return "", err
	}
	return login, nil
}

func (r *orderRepository) Withdraw(ctx context.Context, login, number string, sum float64) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	qtx := r.q.WithTx(tx)

	err = qtx.UploadWithdrawal(ctx, gen.UploadWithdrawalParams{
		Number:    number,
		UserLogin: login,
		Sum: pgtype.Float4{
			Float32: float32(sum),
		},
	})
	if err != nil {
		errr := tx.Rollback(ctx)
		if errr != nil {
			logger.Log.Error("failed to rollback withdrawal transaction")
		}
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return fmt.Errorf("order number is already exist")
		}
		return err
	}

	err = qtx.UpdateUserBalance(ctx, gen.UpdateUserBalanceParams{
		Login: login,
		Current: pgtype.Float4{
			Float32: float32(-sum),
		},
		Withdrawn: pgtype.Float4{
			Float32: float32(sum),
		},
	})

	if err != nil {
		errr := tx.Rollback(ctx)
		if errr != nil {
			logger.Log.Error("failed to rollback withdrawal transaction")
		}

		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *orderRepository) GetUserOrders(ctx context.Context, login string) (*[]models.Order, error) {
	dbOrders, err := r.q.GetUserOrders(ctx, login)
	if err != nil {
		return nil, err
	}

	return dbToModelOrders(&dbOrders), nil
}

func (r *orderRepository) GetOrdersWithStatus(ctx context.Context, status string) (*[]models.Order, error) {
	dbOrders, err := r.q.GetOrdersWithStatus(ctx, pgtype.Text{String: status})
	if err != nil {
		return nil, err
	}

	return dbToModelOrders(&dbOrders), nil
}

func (r *orderRepository) GetUpprocessedOrders(ctx context.Context) (*[]models.Order, error) {
	dbOrders, err := r.q.GetUnprocessedOrders(ctx)
	if err != nil {
		return nil, err
	}

	return dbToModelOrders(&dbOrders), nil
}

func (r *orderRepository) GetUserWithdrawals(ctx context.Context, login string) (*[]models.Withdrawal, error) {
	dbOrders, err := r.q.GetUserWithdrawals(ctx, login)
	if err != nil {
		return nil, err
	}

	return dbToModelWithdrawals(&dbOrders), nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, number, status string) error {
	return r.q.UpdateOrderStatus(ctx, gen.UpdateOrderStatusParams{
		Number: number,
		Status: pgtype.Text{String: status},
	})
}

func (r *orderRepository) OrderProcessed(ctx context.Context, login, number string, accrual float64) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	qtx := r.q.WithTx(tx)

	err = qtx.UpdateOrderAccrual(ctx, gen.UpdateOrderAccrualParams{
		Number: number,
		Accrual: pgtype.Float4{
			Float32: float32(accrual),
		},
	})

	if err != nil {
		errr := tx.Rollback(ctx)
		if errr != nil {
			logger.Log.Error("failed to rollback withdrawal transaction")
		}
		return fmt.Errorf("failed to update accrual. Order number: %s", number)
	}

	err = qtx.UpdateOrderStatus(ctx, gen.UpdateOrderStatusParams{
		Number: number,
		Status: pgtype.Text{String: models.OrderStatusPROCESSED},
	})

	if err != nil {
		errr := tx.Rollback(ctx)
		if errr != nil {
			logger.Log.Error("failed to rollback withdrawal transaction")
		}
		return fmt.Errorf("failed to update order status to PROCESSED. Order number: %s", number)
	}

	if accrual > 0 {
		err = qtx.UpdateUserBalance(ctx, gen.UpdateUserBalanceParams{
			Login: login,
			Current: pgtype.Float4{
				Float32: float32(accrual),
			},
			Withdrawn: pgtype.Float4{
				Float32: float32(0),
			},
		})
	}

	if err != nil {
		errr := tx.Rollback(ctx)
		if errr != nil {
			logger.Log.Panic("failed to rollback withdrawal transaction")
		}

		return fmt.Errorf("failed to update user balance. User login: %s", login)
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *orderRepository) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	order, err := r.q.GetOrderByNumber(ctx, number)
	if err != nil {
		return nil, err
	}

	return dbToModelOrder(&order), nil
}
