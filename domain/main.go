package domain

import (
	"github.com/akeren/go-api-foundry/config"
	"github.com/akeren/go-api-foundry/domain/account"
	"github.com/akeren/go-api-foundry/domain/monitoring"
	"github.com/akeren/go-api-foundry/domain/transaction"
)

func SetupCoreDomain(appConfig *config.ApplicationConfig) {
	// Use factory to create controllers
	monitoringFactory := monitoring.NewMonitoringControllerFactory(appConfig.DB, appConfig.Logger, appConfig.Cache)
	appConfig.RouterService.MountController(monitoringFactory.CreateController())

	accountFactory := account.NewServiceFactory(appConfig.DB, appConfig.Logger)
	appConfig.RouterService.MountController(accountFactory.CreateController())

	transactionFactory := transaction.NewServiceFactory(appConfig.DB, appConfig.Logger)
	appConfig.RouterService.MountController(transactionFactory.CreateController())

}
