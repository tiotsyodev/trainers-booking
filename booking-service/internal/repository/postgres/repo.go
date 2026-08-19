package postgres

import (
	"booker/booking-service/internal/domain"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	*pgxpool.Pool
	TimeoutOperation time.Duration
}

func NewConnPool(ctx context.Context, cfg Config) (Repo, error) {

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	poolCfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return Repo{}, fmt.Errorf("parse connection pool cfg: %w", err)
	}

	dbpool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return Repo{}, fmt.Errorf("unable to create connection pool: %w", err)
	}

	pingTimeOut, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err = dbpool.Ping(pingTimeOut); err != nil {
		dbpool.Close()
		return Repo{}, fmt.Errorf("unable to ping database: %w", err)
	}

	return Repo{Pool: dbpool, TimeoutOperation: cfg.Timeout}, nil
}

var _ domain.Repository = (*Repo)(nil)
