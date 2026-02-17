package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/akeren/go-api-foundry/internal/models"
	apperrors "github.com/akeren/go-api-foundry/pkg/errors"
)

// Transfer funds between accounts
func (s *Service) Transfer(ctx context.Context, req *TransferRequest) (*TransactionResponse, error) {
	if req.Amount <= 0 {
		return nil, apperrors.NewInvalidRequestError("amount must be positive", nil)
	}
	if req.SenderID == req.ReceiverID {
		return nil, apperrors.NewInvalidRequestError("sender and receiver must be different", nil)
	}

	reference := req.Reference

	if reference == "" {
		reference = fmt.Sprintf("%d", time.Now().UnixMilli())
	}

	existingTxn, err := s.repo.FindTransactionByReference(ctx, reference)
	if err != nil {
		return nil, err
	}

	if existingTxn != nil {
		return nil, apperrors.NewConflictError("transaction already exists", nil)
	}

	senderAccount, err := s.repo.FindAccountByID(ctx, req.SenderID)
	if err != nil {
		return nil, err
	}

	receiverAccount, err := s.repo.FindAccountByID(ctx, req.ReceiverID)
	if err != nil {
		return nil, err
	}

	if senderAccount.Currency != receiverAccount.Currency {
		return nil, apperrors.NewInvalidRequestError("sender and receiver must have the same currency", nil)
	}

	amount := int64(req.Amount * 100)

	fmt.Println("amount", amount)
	fmt.Println("senderAccount.Balance", senderAccount.Balance)

	if senderAccount.Balance < amount {
		return nil, apperrors.NewInvalidRequestError("insufficient funds", nil)
	}

	var txn *models.Transaction
	var debitEntry, creditEntry *models.LedgerEntry

	err = s.repo.WithTx(ctx, func(txRepo Repository) error {
		var err error

		txn, err = txRepo.CreateTransaction(ctx, &models.Transaction{
			Reference: reference,
			Type:      "transfer",
			Status:    "completed",
		})
		if err != nil {
			return err
		}

		senderBal, err := txRepo.UpdateBalance(ctx, senderAccount.ID, -amount)
		if err != nil {
			return err
		}
		if senderBal.Balance < 0 {
			return apperrors.NewInvalidRequestError("insufficient funds", nil)
		}

		debitEntry, err = txRepo.CreateLedgerEntry(ctx, &models.LedgerEntry{
			TransactionID:      txn.ID,
			AccountID:          req.SenderID,
			Amount:             amount,
			Direction:          "debit",
			BalanceAfter:       senderBal.Balance,
			LockedBalanceAfter: senderBal.LockedBalance,
		})
		if err != nil {
			return err
		}

		receiverBal, err := txRepo.UpdateBalance(ctx, receiverAccount.ID, amount)
		if err != nil {
			return err
		}

		creditEntry, err = txRepo.CreateLedgerEntry(ctx, &models.LedgerEntry{
			TransactionID:      txn.ID,
			AccountID:          req.ReceiverID,
			Amount:             amount,
			Direction:          "credit",
			BalanceAfter:       receiverBal.Balance,
			LockedBalanceAfter: receiverBal.LockedBalance,
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
