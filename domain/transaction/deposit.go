package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/akeren/go-api-foundry/internal/models"
	apperrors "github.com/akeren/go-api-foundry/pkg/errors"
)

// Deposit funds into an account
func (s *Service) Deposit(ctx context.Context, req *DepositRequest) (*TransactionResponse, error) {
	if req.Amount <= 0 {
		return nil, apperrors.NewInvalidRequestError("amount must be positive", nil)
	}

	var txn *models.Transaction
	var creditEntry, debitEntry *models.LedgerEntry

	err := s.repo.WithTx(ctx, func(txRepo Repository) error {
		var err error

		reference := req.Reference

		if reference == "" {
			reference = fmt.Sprintf("%d", time.Now().UnixMilli())
		}

		txn, err = txRepo.CreateTransaction(ctx, &models.Transaction{
			Reference: reference,
			Type:      "deposit",
			Status:    "completed",
		})
		if err != nil {
			return err
		}

		sysAccount, err := txRepo.FindAccountByType(ctx, "treasury")
		if err != nil {
			return apperrors.NewInternalServerError("system configuration error: external funding account missing", err)
		}

		amountInSubUnits := int64(req.Amount * 100)

		sysBal, err := txRepo.UpdateBalance(ctx, sysAccount.ID, amountInSubUnits)
		if err != nil {
			return err
		}

		debitEntry, err = txRepo.CreateLedgerEntry(ctx, &models.LedgerEntry{
			TransactionID:      txn.ID,
			AccountID:          sysAccount.ID,
			Amount:             amountInSubUnits,
			Direction:          "debit",
			BalanceAfter:       sysBal.Balance,
			LockedBalanceAfter: sysBal.LockedBalance,
		})
		if err != nil {
			return err
		}

		userAccount, err := txRepo.FindAccountByID(ctx, req.AccountID)
		if err != nil {
			return err
		}

		userBal, err := txRepo.UpdateBalance(ctx, userAccount.ID, amountInSubUnits)
		if err != nil {
			return err
		}

		creditEntry, err = txRepo.CreateLedgerEntry(ctx, &models.LedgerEntry{
			TransactionID:      txn.ID,
			AccountID:          req.AccountID,
			Amount:             amountInSubUnits,
			Direction:          "credit",
			BalanceAfter:       userBal.Balance,
			LockedBalanceAfter: userBal.LockedBalance,
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &TransactionResponse{
		ID:        txn.ID,
		Type:      txn.Type,
		Reference: txn.Reference,
		Status:    txn.Status,
		Entries:   []LedgerEntryResponse{ToLedgerEntryResponse(debitEntry), ToLedgerEntryResponse(creditEntry)},
	}, nil
}
