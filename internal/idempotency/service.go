package idempotency

import (
	"context"
	"errors"

	"github.com/akeren/go-api-foundry/internal/log"
	apperrors "github.com/akeren/go-api-foundry/pkg/errors"
	"github.com/akeren/go-api-foundry/sqlc"
)

var (
	ErrKeyProcessing = errors.New("request already in progress")
	ErrKeyMismatch   = errors.New("idempotency key parameter mismatch")
)

type Service struct {
	repo   Repository
	logger *log.Logger
}

func NewService(repo Repository, logger *log.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) CheckIdempotencyKey(ctx context.Context, key string, userID *uint64, path string, params []byte) (*sqlc.IdempotencyKey, error) {
	ik, err := s.repo.GetIdempotencyKey(ctx, key, userID)
	if err == nil {
		if ik.RecoveryPoint.Valid && ik.RecoveryPoint.String == "started" {
			return nil, ErrKeyProcessing
		}

		return ik, nil
	}

	if apperrors.Type(err) != apperrors.ErrorTypeNotFound {
		return nil, err
	}

	ik, err = s.repo.CreateIdempotencyKey(ctx, key, userID, path, params)
	if err != nil {
		// possible race condition
		if apperrors.Type(err) == apperrors.ErrorTypeConflict {
			return nil, ErrKeyProcessing
		}
		return nil, err
	}

	return ik, nil
}

func (s *Service) CompleteRequest(ctx context.Context, key string, userID *uint64, code int, message string, body []byte) error {
	_, err := s.repo.UpdateIdempotencyKeyResponse(ctx, key, userID, code, message, body)
	return err
}

func (s *Service) FailRequest(ctx context.Context, key string, userID *uint64) error {
	return s.repo.DeleteIdempotencyKey(ctx, key, userID)
}
