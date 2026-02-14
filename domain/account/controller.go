package account

import (
	"github.com/akeren/go-api-foundry/config/router"
	"github.com/akeren/go-api-foundry/internal/log"
	apperrors "github.com/akeren/go-api-foundry/pkg/errors"
	"gorm.io/gorm"
)

// NewAccountController creates and returns a versioned RESTController for the account domain
func NewAccountController(db *gorm.DB, logger *log.Logger) *router.RESTController {
	return router.NewVersionedRESTController(
		"AccountController",
		"v1",
		"/account",
		func(rs *router.RouterService, c *router.RESTController) {
			repository := NewAccountRepository(db)
			service := NewAccountService(logger, repository)

			// Register handlers
			rs.AddPostHandler(c, "", createHandler(service))
			rs.AddGetHandler(c, "/:id", getByIDHandler(service))
		},
	)
}

func createHandler(service AccountService) router.HandlerFunction {
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

func getByIDHandler(service AccountService) router.HandlerFunction {
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
