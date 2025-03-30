package repositories

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/morzisorn/gofermart/config"
	"github.com/morzisorn/gofermart/internal/logger"
	"github.com/morzisorn/gofermart/internal/repositories/database"
	gen "github.com/morzisorn/gofermart/internal/repositories/database/generated"
)

var (
	once sync.Once
)

func NewRepository(cfg *config.Config) Repository {
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURI)
	if err != nil {
		logger.Log.Panic(err.Error())
	}

	once.Do(func() {
		err = createTables(db)
		if err != nil {
			logger.Log.Panic(err.Error())
		}
	})

	q := gen.New(db)

	return &DBRepository{
		users:  database.NewUserRepository(q),
		orders: database.NewOrderRepository(q, db),
	}
}

func createTables(db *pgxpool.Pool) error {
	rootDir, err := config.GetProjectRoot()
	if err != nil {
		return err
	}
	filepath := filepath.Join(rootDir, "internal", "repositories", "database", "schema", "schema.sql")

	script, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), string(script))
	if err != nil {
		return err
	}

	return nil
}
