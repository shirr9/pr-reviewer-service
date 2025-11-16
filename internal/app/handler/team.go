package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	teamDto "github.com/shirr9/pr-reviewer-service/internal/app/dto/team"
)

// TeamService defines the interface for team operations.
type TeamService interface {
	AddTeam(ctx context.Context, req teamDto.AddTeamRequest) (*teamDto.AddTeamResponse, error)
	GetTeam(ctx context.Context, teamName string) (*teamDto.GetTeamResponse, error)
	DeactivateTeam(ctx context.Context, teamName string) (*teamDto.DeactivateTeamResponse, error)
}

// TeamHandler handles team related HTTP requests.
type TeamHandler struct {
	service  TeamService
	logger   *slog.Logger
	validate *validator.Validate
}

// NewTeamHandler creates a new TeamHandler.
func NewTeamHandler(
	service TeamService,
	logger *slog.Logger,
	validate *validator.Validate) *TeamHandler {
	if logger == nil {
		logger = slog.Default()
	}
	if validate == nil {
		validate = validator.New()
	}
	return &TeamHandler{
		service:  service,
		logger:   logger,
		validate: validate,
	}
}

// AddTeam add team.
func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	op := "TeamHandler.AddTeam"
	logger := h.logger.With(slog.String("op", op))
	var req teamDto.AddTeamRequest
	if err := decodeAndValidate(r, h.validate, &req); err != nil {
		handleValidationError(w, err, logger)
		return
	}
	response, err := h.service.AddTeam(r.Context(), req)
	if err != nil {
		handleServiceError(w, err, logger)
		return
	}
	sendSuccessResponse(w, http.StatusCreated, response, logger)
}

// GetTeam get team
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	op := "TeamHandler.GetTeam"
	logger := h.logger.With(slog.String("op", op))
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		handleValidationError(w, fmt.Errorf("team_name is required"), logger)
		return
	}
	response, err := h.service.GetTeam(r.Context(), teamName)
	if err != nil {
		handleServiceError(w, err, logger)
		return
	}
	sendSuccessResponse(w, http.StatusOK, response, logger)
}

// DeactivateTeam deactivates all users in a team and reassigns open PRs.
func (h *TeamHandler) DeactivateTeam(w http.ResponseWriter, r *http.Request) {
	op := "TeamHandler.DeactivateTeam"
	logger := h.logger.With(slog.String("op", op))
	var req teamDto.DeactivateTeamRequest
	if err := decodeAndValidate(r, h.validate, &req); err != nil {
		handleValidationError(w, err, logger)
		return
	}
	response, err := h.service.DeactivateTeam(r.Context(), req.TeamName)
	if err != nil {
		handleServiceError(w, err, logger)
		return
	}
	sendSuccessResponse(w, http.StatusOK, response, logger)
}
