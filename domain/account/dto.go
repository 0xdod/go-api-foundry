package account

import (
	"time"

	"github.com/akeren/go-api-foundry/internal/models"
)

// CreateAccountRequest defines the structure for creating a new account entry
type CreateAccountRequest struct {
	Type string `json:"type" binding:"required"`
}

// AccountResponse defines the structure for account responses
type AccountResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Type      string    `json:"type"`
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
		Type: req.Type,
	}
}

// ToAccountResponse converts a models.Account to a AccountResponse
func ToAccountResponse(model *models.Account) AccountResponse {
	if model == nil {
		return AccountResponse{}
	}
	return AccountResponse{
		ID:        uint(model.ID),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Type:      model.Type,
	}
}
