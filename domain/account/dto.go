package account

import (
	"time"

	"github.com/akeren/go-api-foundry/internal/models"
)

// CreateAccountRequest defines the structure for creating a new account entry
type CreateAccountRequest struct {
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Currency string `json:"currency" binding:"required"` // validate currency
	UserID   uint64 `json:"user_id" binding:"required"`
}

// AccountResponse defines the structure for account responses
type AccountResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Currency  string    `json:"currency"`
	UserID    uint64    `json:"user_id"`
	Balance   float64   `json:"balance"` // Cached balance
}

// ========================================
// Mappers
// ========================================

// ToAccountModel converts a CreateAccountRequest to a models.Account
func ToAccountModel(req *CreateAccountRequest) *models.Account {
	if req == nil {
		return nil
	}
	return &models.Account{
		Name:     req.Name,
		Code:     req.Code,
		Currency: req.Currency,
		UserID:   req.UserID,
	}
}

// ToAccountResponse converts a models.Account to a AccountResponse
func ToAccountResponse(model *models.Account) AccountResponse {
	if model == nil {
		return AccountResponse{}
	}

	balanceInUnit := float64(model.Balance) / 100.00

	return AccountResponse{
		ID:        uint(model.ID),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Name:      model.Name,
		Code:      model.Code,
		Currency:  model.Currency,
		UserID:    model.UserID,
		Balance:   balanceInUnit,
	}
}

type BalanceResponse struct {
	Balance float64 `json:"balance"`
}
