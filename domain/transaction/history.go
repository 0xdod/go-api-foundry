package transaction

import (
	"context"
)

// GetHistory retrieves the transaction history for an account
func (s *Service) GetHistory(ctx context.Context, accountID uint64) ([]LedgerEntryResponse, error) {
	_, err := s.repo.FindAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	entries, err := s.repo.GetLedgerEntries(ctx, accountID)
	if err != nil {
		return nil, err
	}

	var response []LedgerEntryResponse
	for _, entry := range entries {
		response = append(response, ToLedgerEntryResponse(entry))
	}

	return response, nil
}
