package transaction

import (
	"strconv"
	"time"

	"github.com/akeren/go-api-foundry/config/router"
	"github.com/akeren/go-api-foundry/internal/idempotency"
	"github.com/akeren/go-api-foundry/internal/log"
	"github.com/akeren/go-api-foundry/internal/postgres"
	apperrors "github.com/akeren/go-api-foundry/pkg/errors"
	"github.com/akeren/go-api-foundry/pkg/ratelimit"
)

// NewController creates and returns a versioned RESTController for the transaction domain
func NewController(db *postgres.DB, logger *log.Logger) *router.RESTController {
	return router.NewVersionedRESTController(
		"TransactionController",
		"v1",
		"/transactions",
		func(rs *router.RouterService, c *router.RESTController) {
			rateLimiter := createTransactionRateLimiter()
			repository := NewRepository(db)
			service := NewService(logger, repository)

			idempotencyRepo := idempotency.NewRepository(db)
			idempotencyService := idempotency.NewService(idempotencyRepo, logger)

			rs.AddPostHandler(c, rateLimiter, "/deposit", idempotency.WithIdempotency(idempotencyService, createDepositHandler(service)))
			rs.AddPostHandler(c, rateLimiter, "/withdraw", idempotency.WithIdempotency(idempotencyService, createWithdrawHandler(service)))
			rs.AddPostHandler(c, rateLimiter, "/transfer", idempotency.WithIdempotency(idempotencyService, createTransferHandler(service)))
			rs.AddGetHandler(c, rateLimiter, "", getHistoryHandler(service))
		},
	)
}

// CreateDeposit godoc
//
//	@Summary		Deposit funds
//	@Description	Deposit funds into an account
//	@Tags			transactions
//	@Accept			json
//	@Produce		json
//	@Param			body			body		DepositRequest	true	"Deposit payload"
//	@Param			idempotency-key	header		string			false	"Idempotency key"
//	@Success		201				{object}	TransactionResponse
//	@Failure		400				{object}	map[string]any
//	@Failure		500				{object}	map[string]any
//	@Router			/v1/transactions/deposit [post]
func createDepositHandler(service *Service) router.HandlerFunction {
	return func(ctx *router.RequestContext) *router.ServiceResult {
		logger := router.GetLogger(ctx)
		var req DepositRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.Error("Failed to bind deposit request", "error", err)

			validationErrors := apperrors.FormatValidationErrors(err, &req)
			if len(validationErrors) > 0 {
				return router.BadRequestResult("Invalid request payload", validationErrors)
			}
			return router.BadRequestResult("Invalid request body", err)
		}

		response, err := service.Deposit(ctx.Request.Context(), &req)
		if err != nil {
			logger.Error("Failed to deposit funds", "error", err)
			return router.ErrorResult(apperrors.HTTPStatusCode(err), apperrors.GetHumanReadableMessage(err), nil)
		}
		return router.CreatedResult(response, "Deposit successful")
	}
}

// CreateWithdraw godoc
//
//	@Summary		Withdraw funds
//	@Description	Withdraw funds from an account
//	@Tags			transactions
//	@Accept			json
//	@Produce		json
//	@Param			body			body		WithdrawRequest	true	"Withdraw payload"
//	@Param			idempotency-key	header		string			false	"Idempotency key"
//	@Success		201				{object}	TransactionResponse
//	@Failure		400				{object}	map[string]any
//	@Failure		500				{object}	map[string]any
//	@Router			/v1/transactions/withdraw [post]
func createWithdrawHandler(service *Service) router.HandlerFunction {
	return func(ctx *router.RequestContext) *router.ServiceResult {
		logger := router.GetLogger(ctx)
		var req WithdrawRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.Error("Failed to bind withdraw request", "error", err)

			validationErrors := apperrors.FormatValidationErrors(err, &req)
			if len(validationErrors) > 0 {
				return router.BadRequestResult("Invalid request payload", validationErrors)
			}
			return router.BadRequestResult("Invalid request body", err)
		}

		response, err := service.Withdraw(ctx.Request.Context(), &req)
		if err != nil {
			logger.Error("Failed to withdraw funds", "error", err)
			return router.ErrorResult(apperrors.HTTPStatusCode(err), apperrors.GetHumanReadableMessage(err), nil)
		}
		return router.CreatedResult(response, "Withdraw successful")
	}
}

// CreateTransfer godoc
//
//	@Summary		Transfer funds
//	@Description	Transfer funds between two accounts
//	@Tags			transactions
//	@Accept			json
//	@Produce		json
//	@Param			body			body		TransferRequest	true	"Transfer payload"
//	@Param			idempotency-key	header		string			false	"Idempotency key"
//	@Success		201				{object}	TransactionResponse
//	@Failure		400				{object}	map[string]any
//	@Failure		500				{object}	map[string]any
//	@Router			/v1/transactions/transfer [post]
func createTransferHandler(service *Service) router.HandlerFunction {
	return func(ctx *router.RequestContext) *router.ServiceResult {
		logger := router.GetLogger(ctx)
		var req TransferRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.Error("Failed to bind transfer request", "error", err)

			validationErrors := apperrors.FormatValidationErrors(err, &req)
			if len(validationErrors) > 0 {
				return router.BadRequestResult("Invalid request payload", validationErrors)
			}
			return router.BadRequestResult("Invalid request body", err)
		}

		response, err := service.Transfer(ctx.Request.Context(), &req)
		if err != nil {
			logger.Error("Failed to transfer funds", "error", err)
			return router.ErrorResult(apperrors.HTTPStatusCode(err), apperrors.GetHumanReadableMessage(err), nil)
		}
		return router.CreatedResult(response, "Transfer successful")
	}
}

// GetHistory godoc
//
//	@Summary		Get transaction history
//	@Description	Get transaction history for an account
//	@Tags			transactions
//	@Accept			json
//	@Produce		json
//	@Param			account_id	query		int	true	"Account ID"
//	@Success		200			{array}		LedgerEntryResponse
//	@Failure		400			{object}	map[string]any
//	@Failure		500			{object}	map[string]any
//	@Router			/v1/transactions [get]
func getHistoryHandler(service *Service) router.HandlerFunction {
	return func(ctx *router.RequestContext) *router.ServiceResult {
		logger := router.GetLogger(ctx)
		accountIDStr := ctx.Query("account_id")
		if accountIDStr == "" {
			return router.BadRequestResult("account_id query parameter is required", nil)
		}

		accountID, err := strconv.ParseUint(accountIDStr, 10, 64)
		if err != nil {
			return router.BadRequestResult("invalid account_id", nil)
		}

		response, err := service.GetHistory(ctx.Request.Context(), accountID)
		if err != nil {
			logger.Error("Failed to get transaction history", "error", err)
			return router.ErrorResult(apperrors.HTTPStatusCode(err), apperrors.GetHumanReadableMessage(err), nil)
		}

		return router.OKResult(response, "Transaction history retrieved successfully")
	}
}

func createTransactionRateLimiter() ratelimit.RateLimiter {
	config := &ratelimit.RateLimitConfig{
		Requests: 100,
		Window:   time.Minute,
		Redis:    nil,
		Logger:   nil,
	}
	return ratelimit.NewRateLimiter(config)
}
