package postgres

import (
	"context"
	"fmt"

	"github.com/akeren/go-api-foundry/internal/log"
	"github.com/akeren/go-api-foundry/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	dsn    string
	logger *log.Logger
	pool   *pgxpool.Pool
	*sqlc.Queries
}

func NewDB(dsn string, logger *log.Logger) *DB {
	return &DB{dsn: dsn, logger: logger}
}

func (c *DB) Connect(ctx context.Context) error {
	c.logger.Info("connecting to database")
	pool, err := pgxpool.New(ctx, c.dsn)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	c.logger.Info("database connected successfully", "host", pool.Config().ConnConfig.Host)
	c.pool = pool
	c.Queries = sqlc.New(pool)

	return nil
}

func (c *DB) Close() {
	c.pool.Close()
}

func (c *DB) WithTx(ctx context.Context, fn func(ctx context.Context, q *sqlc.Queries) error) error {
	tx, err := c.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ctx = SetTx(ctx, tx)
	err = fn(ctx, c.Queries.WithTx(tx))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (c *DB) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}
