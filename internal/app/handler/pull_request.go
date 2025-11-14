package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	prDto "github.com/shirr9/pr-reviewer-service/internal/app/dto/pullrequest"
)

// PullRequestService defines the interface for pull request operations.
type PullRequestService interface {
	CreatePR(ctx context.Context, req prDto.CreatePrRequest) (*prDto.CreatePrResponse, error)
	MergePR(ctx context.Context, req prDto.MergePrRequest) (*prDto.MergePrResponse, error)
	ReassignReviewer(ctx context.Context, req prDto.ReassignReviewerRequest) (*prDto.ReassignReviewerResponse, error)
}

// PullRequestHandler handles pull request related HTTP requests.
type PullRequestHandler struct {
	service  PullRequestService
	logger   *slog.Logger
	validate *validator.Validate
}

// NewPullRequestHandler create new PullRequestHandler.
func NewPullRequestHandler(
	service PullRequestService,
	logger *slog.Logger,
	validate *validator.Validate) *PullRequestHandler {
	return &PullRequestHandler{
		service:  service,
		logger:   logger,
		validate: validate,
	}
}

// CreatePR creates pull request.
func (h *PullRequestHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req prDto.CreatePrRequest
	if err := decodeAndValidate(r, h.validate, &req); err != nil {
		handleValidationError(w, err, h.logger)
		return
	}
	response, err := h.service.CreatePR(r.Context(), req)
	if err != nil {
		handleServiceError(w, err, h.logger)
		return
	}
	sendSuccessResponse(w, http.StatusCreated, response, h.logger)
}

// MergePR merges pull request.
func (h *PullRequestHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req prDto.MergePrRequest
	if err := decodeAndValidate(r, h.validate, &req); err != nil {
		handleValidationError(w, err, h.logger)
		return
	}
	response, err := h.service.MergePR(r.Context(), req)
	if err != nil {
		handleServiceError(w, err, h.logger)
		return
	}
	sendSuccessResponse(w, http.StatusOK, response, h.logger)
}

// ReassignReviewer reassigning a reviewer of pull request.
func (h *PullRequestHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req prDto.ReassignReviewerRequest
	if err := decodeAndValidate(r, h.validate, &req); err != nil {
		handleValidationError(w, err, h.logger)
		return
	}
	response, err := h.service.ReassignReviewer(r.Context(), req)
	if err != nil {
		handleServiceError(w, err, h.logger)
		return
	}
	sendSuccessResponse(w, http.StatusOK, response, h.logger)
}
