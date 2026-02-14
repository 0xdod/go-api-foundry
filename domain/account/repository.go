package account

import (
	"context"

	"github.com/akeren/go-api-foundry/internal/models"
	"github.com/akeren/go-api-foundry/internal/postgres"
)

// Repository defines the data access layer for account domain
type Repository interface {
	// Create persists a new account entry to the database
	Create(ctx context.Context, entry *models.Account) (*models.Account, error)
	// FindByID retrieves a account entry by its unique ID
	FindByID(ctx context.Context, id uint) (*models.Account, error)
}

type repoImpl struct {
	db *postgres.DB
}

// NewRepository creates a new instance of Repository
func NewRepository(db *postgres.DB) Repository {
	return &repoImpl{db: db}
}

// Create persists a new account entry to the database
func (r *repoImpl) Create(ctx context.Context, entry *models.Account) (*models.Account, error) {
	account, err := r.db.Queries.CreateAccount(ctx, entry.Type)
	if err != nil {
		return nil, err
	}

	return &models.Account{
		ID:        uint64(account.ID),
		Type:      account.Type,
		CreatedAt: account.CreatedAt.Time,
		UpdatedAt: account.UpdatedAt.Time,
	}, nil
}

// FindByID retrieves a account entry by its unique ID
func (r *repoImpl) FindByID(ctx context.Context, id uint) (*models.Account, error) {
	account, err := r.db.Queries.FindAccountByID(ctx, int64(id))
	if err != nil {
		return nil, postgres.CheckErrNoRows(err, "account not found")
	}

	return &models.Account{
		ID:        uint64(account.ID),
		Type:      account.Type,
		CreatedAt: account.CreatedAt.Time,
		UpdatedAt: account.UpdatedAt.Time,
	}, nil
}
