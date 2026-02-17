package account

import (
	"context"
	"testing"
	"time"

	"github.com/akeren/go-api-foundry/internal/log"
	"github.com/akeren/go-api-foundry/internal/models"
	database "github.com/akeren/go-api-foundry/internal/postgres"
	apperrors "github.com/akeren/go-api-foundry/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAccountService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	logger := log.NewLoggerWithJSONOutput()
	service := NewService(logger, mockRepo)

	t.Run("successful creation", func(t *testing.T) {
		req := &CreateAccountRequest{
			Name:     "My Savings",
			Code:     "SAV-001",
			Currency: "USD",
			UserID:   1,
		}

		expectedModel := &models.Account{
			Name:     "My Savings",
			Code:     "SAV-001",
			Currency: "USD",
			UserID:   1,
		}

		now := time.Now()
		expectedEntry := &models.Account{
			ID:        1,
			Type:      "user",
			Name:      "My Savings",
			Code:      "SAV-001",
			Currency:  "USD",
			UserID:    1,
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, entry *models.Account) (*models.Account, error) {
				assert.Equal(t, expectedModel.Name, entry.Name)
				assert.Equal(t, expectedModel.Code, entry.Code)
				assert.Equal(t, expectedModel.Currency, entry.Currency)
				assert.Equal(t, expectedModel.UserID, entry.UserID)
				return expectedEntry, nil
			})

		result, err := service.Create(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(expectedEntry.ID), result.ID)
		assert.Equal(t, expectedEntry.Name, result.Name)
	})

	t.Run("empty request", func(t *testing.T) {
		result, err := service.Create(context.Background(), nil)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("repository error", func(t *testing.T) {
		req := &CreateAccountRequest{
			Name:     "My Savings",
			Code:     "SAV-001",
			Currency: "USD",
			UserID:   1,
		}

		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil, apperrors.NewDatabaseError("db error", nil))

		result, err := service.Create(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestAccountService_FindByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	logger := log.NewLoggerWithJSONOutput()
	service := NewService(logger, mockRepo)

	t.Run("successful retrieval", func(t *testing.T) {
		id := uint(1)
		expectedEntry := &models.Account{
			ID:        uint64(id),
			Type:      "savings",
			Name:      "My Savings",
			Code:      "SAV-001",
			Currency:  "USD",
			UserID:    1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockRepo.EXPECT().
			FindByID(gomock.Any(), id).
			Return(expectedEntry, nil)

		result, err := service.FindByID(context.Background(), id)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(id), result.ID)
	})

	t.Run("invalid id", func(t *testing.T) {
		result, err := service.FindByID(context.Background(), 0)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("not found", func(t *testing.T) {
		id := uint(999)
		mockRepo.EXPECT().
			FindByID(gomock.Any(), id).
			Return(nil, database.CheckErrNoRows(apperrors.NewNotFoundError("not found", nil), "account not found"))

		result, err := service.FindByID(context.Background(), id)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestAccountService_GetComputedBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	logger := log.NewLoggerWithJSONOutput()
	service := NewService(logger, mockRepo)

	t.Run("successful balance computation", func(t *testing.T) {
		id := uint(1)
		balance := int64(100)
		expectedBalance := float64(balance) / 100.00
		now := time.Now()
		mockRepo.EXPECT().
			FindByID(gomock.Any(), id).
			Return(&models.Account{
				ID:        uint64(id),
				Type:      "savings",
				Name:      "My Savings",
				Code:      "SAV-001",
				Balance:   balance,
				Currency:  "USD",
				UserID:    1,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil)

		mockRepo.EXPECT().
			ComputeBalance(gomock.Any(), id).
			Return(balance, nil)

		result, err := service.GetComputedBalance(context.Background(), id)

		assert.NoError(t, err)
		assert.Equal(t, expectedBalance, result.Balance)
	})

	t.Run("invalid id", func(t *testing.T) {
		id := uint(0)
		mockRepo.EXPECT().
			FindByID(gomock.Any(), id).
			Return(nil, database.CheckErrNoRows(apperrors.NewNotFoundError("not found", nil), "account not found"))
		result, err := service.GetComputedBalance(context.Background(), id)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("repository error", func(t *testing.T) {
		id := uint(1)
		mockRepo.EXPECT().
			FindByID(gomock.Any(), id).
			Return(nil, apperrors.NewDatabaseError("db error", nil))

		result, err := service.GetComputedBalance(context.Background(), id)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
