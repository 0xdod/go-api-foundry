package monitoring

import (
	"context"

	"github.com/akeren/go-api-foundry/config/router"
	"github.com/akeren/go-api-foundry/internal/log"
	"github.com/akeren/go-api-foundry/internal/postgres"
)

// MonitoringCache defines the cache interface for the monitoring controller factory.
type MonitoringCache interface {
	Ping(ctx context.Context) error
}

type MonitoringControllerFactory interface {
	CreateController() *router.RESTController
}

type DefaultMonitoringControllerFactory struct {
	db     *postgres.DB
	logger *log.Logger
	cache  MonitoringCache
}

func NewMonitoringControllerFactory(db *postgres.DB, logger *log.Logger, cache MonitoringCache) MonitoringControllerFactory {
	return &DefaultMonitoringControllerFactory{
		db:     db,
		logger: logger,
		cache:  cache,
	}
}

func (f *DefaultMonitoringControllerFactory) CreateController() *router.RESTController {
	return NewMonitoringController(f.db, f.logger, f.cache)
}
