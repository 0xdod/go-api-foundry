package account

import (
	"context"

	"github.com/akeren/go-api-foundry/internal/models"
	"github.com/akeren/go-api-foundry/internal/postgres"
	"github.com/akeren/go-api-foundry/sqlc"
)

// Repository defines the data access layer for account domain
type Repository interface {
	// Create persists a new account entry to the database
	Create(ctx context.Context, entry *models.Account) (*models.Account, error)
	// FindByID retrieves a account entry by its unique ID
	FindByID(ctx context.Context, id uint) (*models.Account, error)
	// ComputeBalance calculates the balance from ledger entries
	ComputeBalance(ctx context.Context, id uint) (int64, error)
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
	account, err := r.db.Queries.CreateAccount(ctx, sqlc.CreateAccountParams{
		Name:     postgres.Text(entry.Name),
		Code:     postgres.Text(entry.Code),
		Currency: entry.Currency,
		UserID:   postgres.Int8(int64(entry.UserID)),
	})
	if err != nil {
		return nil, err
	}

	return &models.Account{
		ID:        uint64(account.ID),
		Type:      account.Type,
		Name:      account.Name.String,
		Code:      account.Code.String,
		Currency:  account.Currency,
		UserID:    uint64(account.UserID.Int64),
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
		Name:      account.Name.String,
		Code:      account.Code.String,
		Currency:  account.Currency,
		UserID:    uint64(account.UserID.Int64),
		Balance:   account.Balance,
		CreatedAt: account.CreatedAt.Time,
		UpdatedAt: account.UpdatedAt.Time,
	}, nil
}

// ComputeBalance calculates the balance from ledger entries
func (r *repoImpl) ComputeBalance(ctx context.Context, id uint) (int64, error) {
	balance, err := r.db.Queries.ComputeBalance(ctx, int64(id))
	if err != nil {
		return 0, postgres.CheckErrNoRows(err, "account not found")
	}
	return balance.Balance, nil
}
