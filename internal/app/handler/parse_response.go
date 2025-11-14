package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shirr9/pr-reviewer-service/internal/app/dto"
	domainErrors "github.com/shirr9/pr-reviewer-service/internal/domain/errors"
)

const (
	CodeInternalError = "INTERNAL_ERROR"
	CodeBadRequest    = "BAD_REQUEST"
)

// RespondWithError handles error responses and returns encoding error if any.
func RespondWithError(w http.ResponseWriter, err error) error {
	var appErr *domainErrors.AppError
	if errors.As(err, &appErr) {
		statusCode := mapErrorCodeToHTTPStatus(appErr.Code)
		return RespondJSON(w, statusCode, dto.NewErrorResponse(appErr.Code, appErr.Message))
	}

	if encodeErr := RespondJSON(w, http.StatusInternalServerError,
		dto.NewErrorResponse(CodeInternalError, "internal server error")); encodeErr != nil {
		return encodeErr
	}
	return err
}

// RespondWithCustomError handles custom error responses.
func RespondWithCustomError(w http.ResponseWriter, statusCode int, errResp dto.ErrorResponse) error {
	return RespondJSON(w, statusCode, errResp)
}

// RespondJSON sends a JSON response and returns error if encoding fails.
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}
	return nil
}

// mapErrorCodeToHTTPStatus maps domain error codes to HTTP status codes.
func mapErrorCodeToHTTPStatus(code string) int {
	switch code {
	case domainErrors.CodeNotFound:
		return http.StatusNotFound
	case domainErrors.CodeNotAssigned:
		return http.StatusBadRequest
	case domainErrors.CodeTeamExists, domainErrors.CodePRExists,
		domainErrors.CodePRMerged, domainErrors.CodeNoCandidate:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
