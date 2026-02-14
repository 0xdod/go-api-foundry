package account

import (
	"context"
	"errors"

	"github.com/akeren/go-api-foundry/internal/models"
	apperrors "github.com/akeren/go-api-foundry/pkg/errors"
	"gorm.io/gorm"
)

// AccountRepository defines the data access layer for account domain
type AccountRepository interface {
	// Create persists a new account entry to the database
	Create(ctx context.Context, entry *models.Account) (*models.Account, error)
	// FindByID retrieves a account entry by its unique ID
	FindByID(ctx context.Context, id uint) (*models.Account, error)
}

type accountRepository struct {
	db *gorm.DB
}

// NewAccountRepository creates a new instance of AccountRepository
func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Create(ctx context.Context, entry *models.Account) (*models.Account, error) {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		if isDuplicateKey(err) {
			return nil, apperrors.NewConflictError("account entry already exists", err)
		}
		return nil, apperrors.NewDatabaseError("unable to create account entry", err)
	}
	return entry, nil
}

func (r *accountRepository) FindByID(ctx context.Context, id uint) (*models.Account, error) {
	var entry models.Account
	if err := r.db.WithContext(ctx).First(&entry, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("account entry not found", err)
		}
		return nil, apperrors.NewDatabaseError("failed to fetch account entry", err)
	}
	return &entry, nil
}

// isDuplicateKey checks if the error is a duplicate key constraint violation
func isDuplicateKey(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey) || apperrors.IsDuplicateKeyError(err)
}
