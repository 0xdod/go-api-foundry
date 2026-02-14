package account

import (
	"time"

	"github.com/akeren/go-api-foundry/config/router"
	"github.com/akeren/go-api-foundry/internal/log"
	"github.com/akeren/go-api-foundry/internal/postgres"
	apperrors "github.com/akeren/go-api-foundry/pkg/errors"
	"github.com/akeren/go-api-foundry/pkg/ratelimit"
)

// NewController creates and returns a versioned RESTController for the account domain
func NewController(db *postgres.DB, logger *log.Logger) *router.RESTController {
	return router.NewVersionedRESTController(
		"AccountController",
		"v1",
		"/accounts",
		func(rs *router.RouterService, c *router.RESTController) {
			accountRateLimiter := createAccountRateLimiter()
			repository := NewRepository(db)
			service := NewService(logger, repository)

			// Register handlers
			rs.AddPostHandler(c, accountRateLimiter, "", createAccountHandler(service))
			rs.AddGetHandler(c, accountRateLimiter, "/:id", getAccountByIDHandler(service))
		},
	)
}

// CreateAccount godoc
// @Summary      Create an account
// @Description  create an account
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        body   body      CreateAccountRequest  true  "Account payload"
// @Success      200  {object}  AccountResponse
// @Failure      400  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Failure      500  {object}  map[string]any
// @Router       /v1/accounts [post]
func createAccountHandler(service *Service) router.HandlerFunction {
	return func(ctx *router.RequestContext) *router.ServiceResult {
		logger := router.GetLogger(ctx)

		var req CreateAccountRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.Error("Failed to bind request", "error", err)

			validationErrors := apperrors.FormatValidationErrors(err, &req)
			if len(validationErrors) > 0 {
				return router.BadRequestResult("Invalid request payload", validationErrors)
			}

			return router.BadRequestResult("Invalid request body", nil)
		}

		response, err := service.Create(ctx.Request.Context(), &req)
		if err != nil {
			return router.ErrorResult(
				apperrors.HTTPStatusCode(err),
				apperrors.GetHumanReadableMessage(err),
				nil,
			)
		}

		return router.CreatedResult(response, "Account entry")
	}
}

// GetAccountByID godoc
// @Summary      Get an account
// @Description  get an account by ID
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  AccountResponse
// @Failure      400  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Failure      500  {object}  map[string]any
// @Router       /v1/accounts/{id} [get]
func getAccountByIDHandler(service *Service) router.HandlerFunction {
	return func(ctx *router.RequestContext) *router.ServiceResult {
		id, errResult := router.ParseIDParam(ctx, "id")
		if errResult != nil {
			return errResult
		}

		response, err := service.FindByID(ctx.Request.Context(), id)
		if err != nil {
			return router.ErrorResult(
				apperrors.HTTPStatusCode(err),
				apperrors.GetHumanReadableMessage(err),
				nil,
			)
		}

		return router.OKResult(response, "Account entry retrieved successfully")
	}
}

func createAccountRateLimiter() ratelimit.RateLimiter {
	const accountRequestsPerMinute = 10 // More restrictive than default 100

	config := &ratelimit.RateLimitConfig{
		Requests: accountRequestsPerMinute,
		Window:   time.Minute, // 1 minute window
		Redis:    nil,         // For now, use in-memory (could be enhanced to use Redis)
		Logger:   nil,         // Logger not needed for in-memory limiter
	}

	return ratelimit.NewRateLimiter(config)
}
