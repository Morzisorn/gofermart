package database

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/morzisorn/gofermart/internal/models"
	gen "github.com/morzisorn/gofermart/internal/repositories/database/generated"
)

func pgxFloat4ToFloat64(i pgtype.Float4) (float64, error) {
	if i.Valid {
		return float64(i.Float32), nil
	}

	return 0, fmt.Errorf("invalid float")
}

func pgTimeToTime(pgTime pgtype.Timestamp) (time.Time, error) {
	if pgTime.Valid {
		return pgTime.Time, nil
	}

	return time.Time{}, fmt.Errorf("invalid time")
}

func pgxTextToString(s pgtype.Text) (string, error) {
	if s.Valid {
		return s.String, nil
	}

	return "", fmt.Errorf("invalid string")
}

func dbToModelOrders(dbOrders *[]gen.Order) (*[]models.Order, error) {
	orders := make([]models.Order, len(*dbOrders))

	for i, o := range *dbOrders {
		order, err := dbToModelOrder(&o)
		if err != nil {
			return nil, err
		}

		orders[i] = *order
	}
	return &orders, nil
}

func dbToModelOrder(o *gen.Order) (*models.Order, error) {
	updatedAt, err := pgTimeToTime(o.UploadedAt)
	if err != nil {
		return nil, fmt.Errorf("convert db to model order error: %w", err)
	}

	status, err := pgxTextToString(o.Status)
	if err != nil {
		return nil, fmt.Errorf("convert db to model order error: %w", err)
	}

	accrual, err := pgxFloat4ToFloat64(o.Accrual)
	if err != nil {
		return nil, fmt.Errorf("convert db to model order error: %w", err)
	}

	return &models.Order{
		Number:     o.Number,
		UploadedAt: updatedAt,
		Status:     status,
		Accrual:    accrual,
		UserLogin:  o.UserLogin,
	}, nil
}

func dbToModelWithdrawals(dbOrders *[]gen.Withdrawal) (*[]models.Withdrawal, error) {
	withdrawals := make([]models.Withdrawal, len(*dbOrders))
	for i, o := range *dbOrders {
		processedAt, err := pgTimeToTime(o.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("convert db to model withdrawal error: %w", err)
		}

		sum, err := pgxFloat4ToFloat64(o.Sum)
		if err != nil {
			return nil, fmt.Errorf("convert db to model withdrawal error: %w", err)
		}

		withdrawals[i] = models.Withdrawal{
			Number:      o.Number,
			ProcessedAt: processedAt,
			Sum:         sum,
		}
	}
	return &withdrawals, nil
}
