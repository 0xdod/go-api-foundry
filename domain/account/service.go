package account

import (
	"context"

	"github.com/akeren/go-api-foundry/internal/log"
	apperrors "github.com/akeren/go-api-foundry/pkg/errors"
)

type Service struct {
	logger     *log.Logger
	repository Repository
}

func NewService(logger *log.Logger, repository Repository) *Service {
	return &Service{
		logger:     logger,
		repository: repository,
	}
}

func (s *Service) Create(ctx context.Context, req *CreateAccountRequest) (*AccountResponse, error) {
	logger := log.GetLoggerInstanceFromContext(ctx, s.logger)

	if req == nil {
		logger.Error("Create received empty request")
		return nil, apperrors.NewInvalidRequestError("request cannot be nil", nil)
	}

	model := ToAccountModel(req)
	entry, err := s.repository.Create(ctx, model)
	if err != nil {
		logger.Error("Failed to create account entry", "error", err)
		return nil, err
	}

	response := ToAccountResponse(entry)

	return &response, nil
}

func (s *Service) FindByID(ctx context.Context, id uint) (*AccountResponse, error) {
	logger := log.GetLoggerInstanceFromContext(ctx, s.logger)

	if id == 0 {
		logger.Error("FindByID received invalid ID")
		return nil, apperrors.NewInvalidRequestError("invalid entry ID", nil)
	}

	entry, err := s.repository.FindByID(ctx, id)
	if err != nil {
		logger.Error("Failed to find account entry", "id", id, "error", err)
		return nil, err
	}

	response := ToAccountResponse(entry)
	return &response, nil
}
