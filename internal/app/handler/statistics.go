package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/shirr9/pr-reviewer-service/internal/app/dto/statistics"
)

type StatisticsService interface {
	GetStatistics(ctx context.Context) (*statistics.StatisticsResponse, error)
}

type StatisticsHandler struct {
	service StatisticsService
	log     *slog.Logger
}

func NewStatisticsHandler(service StatisticsService, log *slog.Logger) *StatisticsHandler {
	return &StatisticsHandler{
		service: service,
		log:     log,
	}
}

func (h *StatisticsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.service.GetStatistics(ctx)
	if err != nil {
		h.log.LogAttrs(ctx, slog.LevelError, "failed to get statistics", slog.String("error", err.Error()))
		if encodeErr := RespondWithError(w, err); encodeErr != nil {
			h.log.LogAttrs(ctx, slog.LevelError, "failed to encode error response", slog.String("error", encodeErr.Error()))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.log.LogAttrs(ctx, slog.LevelError, "failed to encode response", slog.String("error", err.Error()))
	}
}
