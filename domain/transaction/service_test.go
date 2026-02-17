package transaction

import (
	"context"
	"testing"

	"github.com/akeren/go-api-foundry/internal/log"
	"github.com/akeren/go-api-foundry/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTransactionService_Deposit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	logger := log.NewLoggerWithJSONOutput()
	service := NewService(logger, mockRepo)

	t.Run("successful deposit", func(t *testing.T) {
		req := &DepositRequest{
			AccountID: 1,
			Amount:    100.00, // $100.00
			Reference: "dep-123",
		}

		treasuryAccount := &models.Account{ID: 999, Type: "treasury"}
		userAccount := &models.Account{ID: 1, Type: "user"}

		txn := &models.Transaction{ID: 1, Reference: "dep-123", Type: "deposit", Status: "completed"}
		sysBal := &models.AccountBalance{AccountID: 999, Balance: 10000}
		userBal := &models.AccountBalance{AccountID: 1, Balance: 10000}

		// Expect WithTx to be called
		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(Repository) error) error {
			return fn(mockRepo)
		})

		// Inside WithTx
		mockRepo.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).Return(txn, nil)
		mockRepo.EXPECT().FindAccountByType(gomock.Any(), "treasury").Return(treasuryAccount, nil)
		// 100.00 * 100 = 10000 cents
		mockRepo.EXPECT().UpdateBalance(gomock.Any(), treasuryAccount.ID, int64(10000)).Return(sysBal, nil)
		mockRepo.EXPECT().CreateLedgerEntry(gomock.Any(), gomock.Any()).Return(&models.LedgerEntry{ID: 1, Amount: 10000}, nil) // Debit Treasury

		mockRepo.EXPECT().FindAccountByID(gomock.Any(), req.AccountID).Return(userAccount, nil)
		mockRepo.EXPECT().UpdateBalance(gomock.Any(), userAccount.ID, int64(10000)).Return(userBal, nil)
		mockRepo.EXPECT().CreateLedgerEntry(gomock.Any(), gomock.Any()).Return(&models.LedgerEntry{ID: 2, Amount: 10000}, nil) // Credit User

		resp, err := service.Deposit(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, txn.ID, resp.ID)
		assert.Equal(t, "completed", resp.Status)
		assert.Len(t, resp.Entries, 2)
	})

	t.Run("invalid amount", func(t *testing.T) {
		req := &DepositRequest{AccountID: 1, Amount: -10, Reference: "invalid"}
		resp, err := service.Deposit(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestTransactionService_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	logger := log.NewLoggerWithJSONOutput()
	service := NewService(logger, mockRepo)

	t.Run("successful withdrawal", func(t *testing.T) {
		req := &WithdrawRequest{
			AccountID: 1,
			Amount:    50.00,
			Reference: "wd-123",
		}

		amountInt := int64(5000)

		userAccount := &models.Account{ID: 1, Type: "user", Balance: 10000} // Sufficient balance

		// 1. Initial Check (Outside Tx)
		mockRepo.EXPECT().FindAccountByID(gomock.Any(), req.AccountID).Return(userAccount, nil)

		// 2. WithTx
		mockRepo.EXPECT().WithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(Repository) error) error {
			return fn(mockRepo)
		})

		// Inside Tx
		txn := &models.Transaction{ID: 2, Reference: "wd-123", Type: "withdraw", Status: "completed"}
		mockRepo.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).Return(txn, nil)

		// Update User Balance (-5000)
		userBalAfter := &models.AccountBalance{AccountID: 1, Balance: 5000}
		mockRepo.EXPECT().UpdateBalance(gomock.Any(), userAccount.ID, -amountInt).Return(userBalAfter, nil)

		// Create Debit Entry (User)
		mockRepo.EXPECT().CreateLedgerEntry(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, entry *models.LedgerEntry) (*models.LedgerEntry, error) {
			assert.Equal(t, "debit", entry.Direction)
			assert.Equal(t, userAccount.ID, getUint64(entry.AccountID))
			return &models.LedgerEntry{ID: 3, Direction: "debit", Amount: amountInt}, nil
		})

		// Find Treasury
		treasuryAccount := &models.Account{ID: 999, Type: "treasury"}
		mockRepo.EXPECT().FindAccountByType(gomock.Any(), "treasury").Return(treasuryAccount, nil)

		// Update Treasury Balance (-5000)
		treasuryBalAfter := &models.AccountBalance{AccountID: 999, Balance: 5000}
		mockRepo.EXPECT().UpdateBalance(gomock.Any(), treasuryAccount.ID, -amountInt).Return(treasuryBalAfter, nil)

		// Create Credit Entry (Treasury)
		mockRepo.EXPECT().CreateLedgerEntry(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, entry *models.LedgerEntry) (*models.LedgerEntry, error) {
			assert.Equal(t, "credit", entry.Direction)
			assert.Equal(t, treasuryAccount.ID, getUint64(entry.AccountID))
			return &models.LedgerEntry{ID: 4, Direction: "credit", Amount: amountInt}, nil
		})

		resp, err := service.Withdraw(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, txn.ID, resp.ID)
	})

	t.Run("insufficient funds check", func(t *testing.T) {
		req := &WithdrawRequest{AccountID: 1, Amount: 100.00, Reference: "wd-fail"}
		userAccount := &models.Account{ID: 1, Balance: 5000} // $50.00 only

		mockRepo.EXPECT().FindAccountByID(gomock.Any(), req.AccountID).Return(userAccount, nil)

		resp, err := service.Withdraw(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// Helper to handle uint64 type flexibility in tests
func getUint64(v interface{}) uint64 {
	switch val := v.(type) {
	case uint64:
		return val
	case int64:
		return uint64(val)
	case int:
		return uint64(val)
	default:
		return 0
	}
}
