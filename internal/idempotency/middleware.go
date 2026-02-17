package idempotency

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"

	"github.com/akeren/go-api-foundry/config/router"
)

// WithIdempotency wraps a handler with idempotency logic
func WithIdempotency(svc *Service, handler router.HandlerFunction) router.HandlerFunction {
	return func(ctx *router.RequestContext) *router.ServiceResult {
		key := ctx.GetHeader("Idempotency-Key")
		if key == "" {
			return handler(ctx)
		}

		path := ctx.Request.URL.Path

		var bodyCopy bytes.Buffer
		if _, err := io.Copy(&bodyCopy, ctx.Request.Body); err != nil && err != io.EOF {
			svc.logger.Error("Failed to process idempotency key", "error", err)
			return router.InternalServerErrorResult("Failed to process idempotency key")
		}

		ctx.Request.Body = io.NopCloser(&bodyCopy)

		var userID *uint64

		uid := ctx.GetUint64("user_id")
		if uid > 0 {
			userID = &uid
		}

		// no auth yet
		if userID == nil {
			defaultUserID := uint64(1)
			userID = &defaultUserID
		}

		ik, err := svc.CheckIdempotencyKey(ctx.Request.Context(), key, userID, path, bodyCopy.Bytes())
		if err != nil {
			if err == ErrKeyProcessing {
				return router.ConflictResult("Request already in progress using this idempotency key")
			}
			svc.logger.Error("Failed to process idempotency key", "error", err)
			return router.InternalServerErrorResult("Failed to process idempotency key")
		}

		if ik != nil && ik.RecoveryPoint.Valid && ik.RecoveryPoint.String == "completed" {
			var b1 any
			var b2 any
			if err := json.Unmarshal(bodyCopy.Bytes(), &b1); err != nil {
				svc.logger.Error("Failed to process idempotency key", "error", err)
			}
			if err := json.Unmarshal(ik.RequestParams, &b2); err != nil {
				svc.logger.Error("Failed to process idempotency key", "error", err)
			}

			if !reflect.DeepEqual(b1, b2) {
				return router.BadRequestResult("Payload mismtach", nil)
			}

			message := ik.Message.String

			if message == "" {
				message = "Cached response"
			}

			return &router.ServiceResult{
				StatusCode: int(ik.ResponseCode.Int32),
				Data:       json.RawMessage(ik.ResponseBody),
				Message:    message,
			}
		}

		result := handler(ctx)

		if result.StatusCode >= 200 && result.StatusCode < 500 {
			bodyBytes, _ := json.Marshal(result.Data)
			if err := svc.CompleteRequest(ctx.Request.Context(), key, userID, result.StatusCode, result.Message, bodyBytes); err != nil {
				svc.logger.Error("Failed to complete idempotency key", "error", err)
			}
		} else {
			if err := svc.FailRequest(ctx.Request.Context(), key, userID); err != nil {
				svc.logger.Error("Failed to fail idempotency key", "error", err)
			}
		}

		return result
	}
}
