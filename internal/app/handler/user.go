package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	userDto "github.com/shirr9/pr-reviewer-service/internal/app/dto/user"
)

// UserService defines the interface for user operations.
type UserService interface {
	SetIsActive(ctx context.Context, req userDto.SetIsActiveRequest) (*userDto.SetIsActiveResponse, error)
	GetReview(ctx context.Context, userID string) (*userDto.GetReviewResponse, error)
}

// UserHandler handles user related HTTP requests.
type UserHandler struct {
	service  UserService
	logger   *slog.Logger
	validate *validator.Validate
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(
	service UserService,
	logger *slog.Logger,
	validate *validator.Validate) *UserHandler {
	return &UserHandler{
		service:  service,
		logger:   logger,
		validate: validate,
	}
}

// SetIsActive handles setIsActive request.
func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req userDto.SetIsActiveRequest
	if err := decodeAndValidate(r, h.validate, &req); err != nil {
		handleValidationError(w, err, h.logger)
		return
	}
	response, err := h.service.SetIsActive(r.Context(), req)
	if err != nil {
		handleServiceError(w, err, h.logger)
		return
	}
	sendSuccessResponse(w, http.StatusOK, response, h.logger)
}

// GetReview handles getReview request.
func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		handleValidationError(w, fmt.Errorf("user_id is required"), h.logger)
		return
	}
	response, err := h.service.GetReview(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err, h.logger)
		return
	}
	sendSuccessResponse(w, http.StatusOK, response, h.logger)
}
