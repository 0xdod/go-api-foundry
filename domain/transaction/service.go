package transaction

import (
	"github.com/akeren/go-api-foundry/internal/log"
)

type Service struct {
	logger *log.Logger
	repo   Repository
}

func NewService(logger *log.Logger, repo Repository) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}
