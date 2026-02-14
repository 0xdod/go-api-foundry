package account

import (
	"github.com/akeren/go-api-foundry/internal/models"
	"github.com/akeren/go-api-foundry/pkg/constants"
)

// CreateAccountRequest defines the structure for creating a new account entry
type CreateAccountRequest struct {
	// Add your request fields here with validation tags
	// Example: Name string `json:"name" binding:"required"`
}

// AccountResponse defines the structure for account responses
type AccountResponse struct {
	ID        uint   `json:"id"`
	CreatedAt string `json:"created_at"`
	// Add your response fields here
	// Example: Name string `json:"name"`
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
		// Map request fields to model fields
		// Example: Name: req.Name,
	}
}

// ToAccountResponse converts a models.Account to a AccountResponse
func ToAccountResponse(model *models.Account) AccountResponse {
	if model == nil {
		return AccountResponse{}
	}
	return AccountResponse{
		ID:        model.ID,
		CreatedAt: model.CreatedAt.Format(constants.RFC3339DateTimeFormat),
		// Map model fields to response fields
		// Example: Name: model.Name,
	}
}
