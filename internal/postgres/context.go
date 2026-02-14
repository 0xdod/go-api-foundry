package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type ctxKey struct{}

var txKey ctxKey

func SetTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func GetTx(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(txKey).(pgx.Tx)
	return tx
}
