package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/shirr9/pr-reviewer-service/internal/app/dto"
)

// decodeAndValidate decode and validate request body.
func decodeAndValidate(r *http.Request, v *validator.Validate, target interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return err
	}
	if err := v.Struct(target); err != nil {
		return err
	}
	return nil
}

// handleValidationError handles validation error and logs it.
func handleValidationError(w http.ResponseWriter, err error, logger *slog.Logger) {
	respErr := RespondWithCustomError(w, http.StatusBadRequest,
		dto.NewErrorResponse(CodeBadRequest, err.Error()))
	if respErr != nil {
		logger.Error("failed to send validation error response",
			slog.String("error", respErr.Error()))
	}
}

// handleServiceError handles service error and logs it.
func handleServiceError(w http.ResponseWriter, err error, logger *slog.Logger) {
	if respErr := RespondWithError(w, err); respErr != nil {
		logger.Error("unexpected error in handler",
			slog.String("error", err.Error()),
			slog.String("response_error", respErr.Error()),
		)
	}
}

// sendSuccessResponse returns success response and logs error if occurs.
func sendSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}, logger *slog.Logger) {
	if err := RespondJSON(w, statusCode, data); err != nil {
		logger.Error("failed to send success response",
			slog.String("error", err.Error()),
		)
	}
}
