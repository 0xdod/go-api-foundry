package transaction

import (
	"context"
	"strconv"

	"github.com/akeren/go-api-foundry/internal/models"
	"github.com/akeren/go-api-foundry/internal/postgres"
	"github.com/akeren/go-api-foundry/sqlc"
)

type Repository interface {
	// CreateTransaction creates a new transaction record
	CreateTransaction(ctx context.Context, model *models.Transaction) (*models.Transaction, error)
	// FindTransactionByReference
	FindTransactionByReference(ctx context.Context, reference string) (*models.Transaction, error)
	// CreateLedgerEntry creates a new ledger entry
	CreateLedgerEntry(ctx context.Context, entry *models.LedgerEntry) (*models.LedgerEntry, error)
	// UpdateBalance updates the account balance atomically
	UpdateBalance(ctx context.Context, accountID uint64, amount int64) (*models.AccountBalance, error)
	// GetBalance retrieves the current balance for an account
	GetBalance(ctx context.Context, accountID uint64) (*models.AccountBalance, error)
	// GetLedgerEntries retrieves ledger entries for a specific account
	GetLedgerEntries(ctx context.Context, accountID uint64) ([]*models.LedgerEntry, error)
	// FindAccountByType finds an account by its type (e.g., 'external_funding', 'treasury')
	FindAccountByType(ctx context.Context, accountType string) (*models.Account, error)
	// FindAccountByID finds an account by its ID
	FindAccountByID(ctx context.Context, accountID uint64) (*models.Account, error)
	// WithTx runs a function within a database transaction
	WithTx(ctx context.Context, fn func(repo Repository) error) error
}

type repoImpl struct {
	db *postgres.DB
	q  *sqlc.Queries
}

func NewRepository(db *postgres.DB) Repository {
	return &repoImpl{
		db: db,
		q:  db.Queries,
	}
}

func (r *repoImpl) WithTx(ctx context.Context, fn func(Repository) error) error {
	return r.db.WithTx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		txRepo := &repoImpl{
			db: r.db,
			q:  q,
		}
		return fn(txRepo)
	})
}

func (r *repoImpl) CreateTransaction(ctx context.Context, model *models.Transaction) (*models.Transaction, error) {
	tx, err := r.q.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		Reference: postgres.Text(model.Reference),
		Status:    model.Status,
		Type:      model.Type,
	})
	if err != nil {
		return nil, postgres.CheckErrUniqueViolation(err, "transaction reference already exists")
	}

	return &models.Transaction{
		ID:        uint64(tx.ID),
		Reference: tx.Reference.String,
		Type:      tx.Type,
		Status:    tx.Status,
		CreatedAt: tx.CreatedAt.Time,
		UpdatedAt: tx.UpdatedAt.Time,
	}, nil
}

func (r *repoImpl) FindTransactionByReference(ctx context.Context, reference string) (*models.Transaction, error) {
	tx, err := r.q.FindTransactionByReference(ctx, postgres.Text(reference))
	if err != nil {
		return nil, postgres.CheckErrNoRows(err, "transaction not found")
	}

	return &models.Transaction{
		ID:        uint64(tx.ID),
		Reference: tx.Reference.String,
		Type:      tx.Type,
		Status:    tx.Status,
		CreatedAt: tx.CreatedAt.Time,
		UpdatedAt: tx.UpdatedAt.Time,
	}, nil
}

func (r *repoImpl) CreateLedgerEntry(ctx context.Context, entry *models.LedgerEntry) (*models.LedgerEntry, error) {
	le, err := r.q.CreateLedgerEntry(ctx, sqlc.CreateLedgerEntryParams{
		TransactionID:      int64(entry.TransactionID),
		AccountID:          int64(entry.AccountID),
		Amount:             entry.Amount,
		Direction:          entry.Direction,
		BalanceAfter:       entry.BalanceAfter,
		LockedBalanceAfter: entry.LockedBalanceAfter,
	})
	if err != nil {
		return nil, err
	}

	return &models.LedgerEntry{
		ID:                 uint64(le.ID),
		TransactionID:      uint64(le.TransactionID),
		AccountID:          uint64(le.AccountID),
		Amount:             le.Amount,
		Direction:          le.Direction,
		BalanceAfter:       le.BalanceAfter,
		LockedBalanceAfter: le.LockedBalanceAfter,
		CreatedAt:          le.CreatedAt.Time,
	}, nil
}

func (r *repoImpl) UpdateBalance(ctx context.Context, accountID uint64, amount int64) (*models.AccountBalance, error) {
	bal, err := r.q.ComputeBalance(ctx, int64(accountID))

	if err != nil {
		return nil, err
	}

	bal, err = r.q.SaveBalance(ctx, sqlc.SaveBalanceParams{
		AccountID: int64(accountID),
		Balance:   bal.Balance + amount,
	})

	if err != nil {
		return nil, postgres.CheckErrCheckViolation(err, "insufficient funds")
	}

	return &models.AccountBalance{
		AccountID:     uint64(bal.AccountID),
		Balance:       bal.Balance,
		LockedBalance: bal.LockedBalance,
	}, nil
}

func (r *repoImpl) GetBalance(ctx context.Context, accountID uint64) (*models.AccountBalance, error) {
	bal, err := r.q.GetBalance(ctx, int64(accountID))
	if err != nil {
		return nil, postgres.CheckErrNoRows(err, "balance not found")
	}

	return &models.AccountBalance{
		AccountID:     uint64(bal.AccountID),
		Balance:       bal.Balance,
		LockedBalance: bal.LockedBalance,
	}, nil
}

func (r *repoImpl) GetLedgerEntries(ctx context.Context, accountID uint64) ([]*models.LedgerEntry, error) {
	entries, err := r.q.FindLedgerEntriesByAccountID(ctx, int64(accountID))
	if err != nil {
		return nil, err
	}

	var result []*models.LedgerEntry
	for _, le := range entries {
		result = append(result, &models.LedgerEntry{
			ID:                 uint64(le.ID),
			TransactionID:      uint64(le.TransactionID),
			Reference:          le.Reference.String,
			Type:               le.Type,
			Status:             le.Status,
			AccountID:          uint64(le.AccountID),
			Amount:             le.Amount,
			Direction:          le.Direction,
			BalanceAfter:       le.BalanceAfter,
			LockedBalanceAfter: le.LockedBalanceAfter,
			CreatedAt:          le.CreatedAt.Time,
		})
	}
	return result, nil
}

func (r *repoImpl) FindAccountByType(ctx context.Context, accountType string) (*models.Account, error) {
	acc, err := r.q.FindAccountByType(ctx, accountType)
	if err != nil {
		return nil, postgres.CheckErrNoRows(err, "system account not found")
	}

	return &models.Account{
		ID:        uint64(acc.ID),
		Type:      acc.Type,
		Name:      acc.Name.String,
		Code:      acc.Code.String,
		Currency:  acc.Currency,
		Balance:   acc.Balance,
		UserID:    uint64(acc.UserID.Int64),
		CreatedAt: acc.CreatedAt.Time,
		UpdatedAt: acc.UpdatedAt.Time,
	}, nil
}

func (r *repoImpl) FindAccountByID(ctx context.Context, accountID uint64) (*models.Account, error) {
	acc, err := r.q.FindAccountByID(ctx, int64(accountID))
	if err != nil {
		return nil, postgres.CheckErrNoRows(err, "account "+strconv.FormatUint(accountID, 10)+" not found")
	}

	return &models.Account{
		ID:        uint64(acc.ID),
		Type:      acc.Type,
		Name:      acc.Name.String,
		Code:      acc.Code.String,
		Currency:  acc.Currency,
		Balance:   acc.Balance,
		UserID:    uint64(acc.UserID.Int64),
		CreatedAt: acc.CreatedAt.Time,
		UpdatedAt: acc.UpdatedAt.Time,
	}, nil
}
