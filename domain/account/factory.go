package account

import (
	"github.com/akeren/go-api-foundry/config/router"
	"github.com/akeren/go-api-foundry/internal/log"
	"github.com/akeren/go-api-foundry/internal/postgres"
)

type ServiceFactory interface {
	CreateService() *Service
	CreateController() *router.RESTController
}

type DefaultServiceFactory struct {
	db     *postgres.DB
	logger *log.Logger
}

func NewServiceFactory(db *postgres.DB, logger *log.Logger) ServiceFactory {
	return &DefaultServiceFactory{
		db:     db,
		logger: logger,
	}
}

func (f *DefaultServiceFactory) CreateService() *Service {
	repository := NewRepository(f.db)
	return NewService(f.logger, repository)
}

func (f *DefaultServiceFactory) CreateController() *router.RESTController {
	return NewController(f.db, f.logger)
}
