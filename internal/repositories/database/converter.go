package database

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/morzisorn/gofermart/internal/logger"
	"github.com/morzisorn/gofermart/internal/models"
	gen "github.com/morzisorn/gofermart/internal/repositories/database/generated"
)

func pgxFloat4ToFloat64(i pgtype.Float4) float64 {
	if i.Valid {
		return float64(i.Float32)
	}
	return 0
}

func pgTimeToTime(pgTime pgtype.Timestamp) time.Time {
	if pgTime.Valid {
		return pgTime.Time
	}	
	logger.Log.Panic("Invalid time")
	return time.Time{}
}

func pgxTextToString(s pgtype.Text) string {
	if s.Valid {
		return s.String
	}
	logger.Log.Panic("Invalid status")
	return ""
}

func dbToModelOrders(dbOrders *[]gen.Order) *[]models.Order {
	orders := make([]models.Order, len(*dbOrders))
	for i, o := range *dbOrders {
		orders[i] = models.Order{
			Number:     o.Number,
			UploadedAt: pgTimeToTime(o.UploadedAt),
			Status:     pgxTextToString(o.Status),
			Accrual:    pgxFloat4ToFloat64(o.Accrual),
		}
	}
	return &orders
}

func dbToModelWithdrawals(dbOrders *[]gen.Withdrawal) *[]models.Withdrawal {
	withdrawals := make([]models.Withdrawal, len(*dbOrders))
	for i, o := range *dbOrders {
		withdrawals[i] = models.Withdrawal{
			Number:      o.Number,
			ProcessedAt: pgTimeToTime(o.ProcessedAt),
			Sum:         pgxFloat4ToFloat64(o.Sum),
		}
	}
	return &withdrawals
}

