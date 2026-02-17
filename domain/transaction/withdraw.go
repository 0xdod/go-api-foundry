package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/akeren/go-api-foundry/internal/models"
	apperrors "github.com/akeren/go-api-foundry/pkg/errors"
)

// Withdraw funds from an account
func (s *Service) Withdraw(ctx context.Context, req *WithdrawRequest) (*TransactionResponse, error) {
	if req.Amount <= 0 {
		return nil, apperrors.NewInvalidRequestError("amount must be positive", nil)
	}

	amount := int64(req.Amount * 100)

	userAccount, err := s.repo.FindAccountByID(ctx, req.AccountID)
	if err != nil {
		return nil, err
	}

	if userAccount.Balance < amount {
		return nil, apperrors.NewInvalidRequestError("insufficient funds", nil)
	}

	var txn *models.Transaction
	var creditEntry, debitEntry *models.LedgerEntry

	err = s.repo.WithTx(ctx, func(txRepo Repository) error {
		var err error

		reference := req.Reference

		if reference == "" {
			reference = fmt.Sprintf("%d", time.Now().UnixMilli())
		}

		txn, err = txRepo.CreateTransaction(ctx, &models.Transaction{
			Reference: reference,
			Type:      "withdraw",
			Status:    "completed",
		})
		if err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		userBal, err := txRepo.UpdateBalance(ctx, userAccount.ID, -amount)
		if err != nil {
			return fmt.Errorf("failed to update user account balance: %w", err)
		}

		if userBal.Balance < 0 {
			return apperrors.NewInvalidRequestError("insufficient funds", nil)
		}

		debitEntry, err = txRepo.CreateLedgerEntry(ctx, &models.LedgerEntry{
			TransactionID:      txn.ID,
			AccountID:          req.AccountID,
			Amount:             amount,
			Direction:          "debit",
			BalanceAfter:       userBal.Balance,
			LockedBalanceAfter: userBal.LockedBalance,
		})
		if err != nil {
			return err
		}

		sysAccount, err := txRepo.FindAccountByType(ctx, "treasury")
		if err != nil {
			return apperrors.NewInternalServerError("system configuration error: external funding account missing", err)
		}

		sysBal, err := txRepo.UpdateBalance(ctx, sysAccount.ID, -amount)
		if err != nil {
			return fmt.Errorf("failed to update system account balance: %w", err)
		}

		creditEntry, err = txRepo.CreateLedgerEntry(ctx, &models.LedgerEntry{
			TransactionID:      txn.ID,
			AccountID:          sysAccount.ID,
			Amount:             amount,
			Direction:          "credit",
			BalanceAfter:       sysBal.Balance,
			LockedBalanceAfter: sysBal.LockedBalance,
		})
		if err != nil {
			return fmt.Errorf("failed to create ledger entry: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &TransactionResponse{
		ID:        txn.ID,
		Reference: txn.Reference,
		Status:    txn.Status,
		Type:      txn.Type,
		Entries:   []LedgerEntryResponse{ToLedgerEntryResponse(debitEntry), ToLedgerEntryResponse(creditEntry)},
	}, nil
}
