package transaction

import "github.com/akeren/go-api-foundry/internal/models"

type DepositRequest struct {
	AccountID uint64  `json:"account_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Reference string  `json:"reference" binding:"required"`
}

type WithdrawRequest struct {
	AccountID uint64  `json:"account_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Reference string  `json:"reference" binding:"required"`
}

type TransferRequest struct {
	SenderID   uint64  `json:"sender_id" binding:"required"`
	ReceiverID uint64  `json:"receiver_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Reference  string  `json:"reference" binding:"required"`
}

type TransactionResponse struct {
	ID        uint64                `json:"id"`
	Type      string                `json:"type"`
	Reference string                `json:"reference"`
	Status    string                `json:"status"`
	Entries   []LedgerEntryResponse `json:"entries"`
}

type LedgerEntryResponse struct {
	ID            uint64  `json:"id"`
	TransactionID uint64  `json:"transaction_id"`
	Reference     string  `json:"reference,omitempty"`
	Type          string  `json:"type,omitempty"`
	Status        string  `json:"status,omitempty"`
	AccountID     uint64  `json:"account_id"`
	Amount        float64 `json:"amount"`
	Direction     string  `json:"direction"`
	BalanceAfter  float64 `json:"balance_after"`
}

func ToLedgerEntryResponse(entry *models.LedgerEntry) LedgerEntryResponse {
	amtInFloat := float64(entry.Amount) / 100.00
	balanceAfterInFloat := float64(entry.BalanceAfter) / 100.00

	return LedgerEntryResponse{
		ID:            entry.ID,
		AccountID:     entry.AccountID,
		TransactionID: entry.TransactionID,
		Amount:        amtInFloat,
		Reference:     entry.Reference,
		Type:          entry.Type,
		Status:        entry.Status,
		Direction:     entry.Direction,
		BalanceAfter:  balanceAfterInFloat,
	}
}
