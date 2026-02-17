package idempotency

import (
	"context"

	"github.com/akeren/go-api-foundry/internal/postgres"
	"github.com/akeren/go-api-foundry/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type Repository interface {
	GetIdempotencyKey(ctx context.Context, key string, userID *uint64) (*sqlc.IdempotencyKey, error)
	CreateIdempotencyKey(ctx context.Context, key string, userID *uint64, path string, params []byte) (*sqlc.IdempotencyKey, error)
	UpdateIdempotencyKeyResponse(ctx context.Context, key string, userID *uint64, code int, message string, body []byte) (*sqlc.IdempotencyKey, error)
	DeleteIdempotencyKey(ctx context.Context, key string, userID *uint64) error
}

type repoImpl struct {
	db *postgres.DB
}

func NewRepository(db *postgres.DB) Repository {
	return &repoImpl{db: db}
}

func (r *repoImpl) GetIdempotencyKey(ctx context.Context, key string, userID *uint64) (*sqlc.IdempotencyKey, error) {
	userIdVal := pgtype.Int8{}
	if userID != nil {
		userIdVal = pgtype.Int8{Int64: int64(*userID), Valid: true}
	} else {
		userIdVal = pgtype.Int8{Valid: false}
	}

	ik, err := r.db.Queries.GetIdempotencyKey(ctx, sqlc.GetIdempotencyKeyParams{
		Key:    key,
		UserID: userIdVal,
	})
	if err != nil {
		return nil, postgres.CheckErrNoRows(err, "idempotency key not found")
	}
	return &ik, nil
}

func (r *repoImpl) CreateIdempotencyKey(ctx context.Context, key string, userID *uint64, path string, params []byte) (*sqlc.IdempotencyKey, error) {
	userIdVal := pgtype.Int8{}
	if userID != nil {
		userIdVal = pgtype.Int8{Int64: int64(*userID), Valid: true}
	} else {
		userIdVal = pgtype.Int8{Valid: false}
	}

	ik, err := r.db.Queries.CreateIdempotencyKey(ctx, sqlc.CreateIdempotencyKeyParams{
		Key:           key,
		UserID:        userIdVal,
		ResourcePath:  path,
		RequestParams: params,
	})
	if err != nil {
		return nil, postgres.CheckErrUniqueViolation(err, "idempotency key already exists")
	}
	return &ik, nil
}

func (r *repoImpl) UpdateIdempotencyKeyResponse(ctx context.Context, key string, userID *uint64, code int, message string, body []byte) (*sqlc.IdempotencyKey, error) {
	userIdVal := pgtype.Int8{}
	if userID != nil {
		userIdVal = pgtype.Int8{Int64: int64(*userID), Valid: true}
	} else {
		userIdVal = pgtype.Int8{Valid: false}
	}

	ik, err := r.db.Queries.UpdateIdempotencyKeyResponse(ctx, sqlc.UpdateIdempotencyKeyResponseParams{
		Key:          key,
		UserID:       userIdVal,
		ResponseCode: pgtype.Int4{Int32: int32(code), Valid: true},
		Message:      pgtype.Text{String: message, Valid: true},
		ResponseBody: body,
	})
	if err != nil {
		return nil, err
	}
	return &ik, nil
}

func (r *repoImpl) DeleteIdempotencyKey(ctx context.Context, key string, userID *uint64) error {
	userIdVal := pgtype.Int8{}
	if userID != nil {
		userIdVal = pgtype.Int8{Int64: int64(*userID), Valid: true}
	} else {
		userIdVal = pgtype.Int8{Valid: false}
	}

	err := r.db.Queries.DeleteIdempotencyKey(ctx, sqlc.DeleteIdempotencyKeyParams{
		Key:    key,
		UserID: userIdVal,
	})
	if err != nil {
		return err
	}
	return nil
}
