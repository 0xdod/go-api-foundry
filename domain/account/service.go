package account

import (
	"context"

	"github.com/akeren/go-api-foundry/internal/log"
	apperrors "github.com/akeren/go-api-foundry/pkg/errors"
)

// AccountService defines the business logic layer for account domain
type AccountService interface {
	// Create creates a new account entry based on the provided request
	Create(ctx context.Context, req *CreateAccountRequest) (*AccountResponse, error)

	// FindByID retrieves a account entry by its unique ID
	FindByID(ctx context.Context, id uint) (*AccountResponse, error)
}

type accountService struct {
	logger     *log.Logger
	repository AccountRepository
}

func NewAccountService(logger *log.Logger, repository AccountRepository) AccountService {
	return &AccountService{
		logger:     logger,
		repository: repository,
	}
}

func (s *accountService) Create(ctx context.Context, req *CreateAccountRequest) (*AccountResponse, error) {
	logger := log.GetLoggerInstanceFromContext(ctx, s.logger)

	if req == nil {
		logger.Error("Create received empty request")
		return nil, apperrors.NewInvalidRequestError("request cannot be nil", nil)
	}

	// Add business validation logic here

	model := ToAccountModel(req)
	entry, err := s.repository.Create(ctx, model)
	if err != nil {
		logger.Error("Failed to create account entry", "error", err)
		return nil, err
	}

	response := ToAccountResponse(entry)
	return &response, nil
}

func (s *accountService) FindByID(ctx context.Context, id uint) (*AccountResponse, error) {
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
