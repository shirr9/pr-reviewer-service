package service

import (
	"context"
	"log/slog"

	"github.com/shirr9/pr-reviewer-service/internal/app/dto/statistics"
	"github.com/shirr9/pr-reviewer-service/internal/domain/models"
)

type StatisticsUserRepository interface {
	GetAllUsers(ctx context.Context) ([]*models.User, error)
}

type StatisticsPRRepository interface {
	GetAllPRs(ctx context.Context) ([]*models.PullRequest, error)
}

type StatisticsReviewerRepository interface {
	GetReviewers(ctx context.Context, prID string) ([]string, error)
	GetAllReviewerCounts(ctx context.Context) (map[string]int, error)
	GetPRsByReviewer(ctx context.Context, reviewerID string) ([]string, error)
}

type StatisticsService struct {
	userRepo     StatisticsUserRepository
	prRepo       StatisticsPRRepository
	reviewerRepo StatisticsReviewerRepository
	log          *slog.Logger
}

func NewStatisticsService(
	userRepo StatisticsUserRepository,
	prRepo StatisticsPRRepository,
	reviewerRepo StatisticsReviewerRepository,
	log *slog.Logger,
) *StatisticsService {
	if log == nil {
		log = slog.Default()
	}
	return &StatisticsService{
		userRepo:     userRepo,
		prRepo:       prRepo,
		reviewerRepo: reviewerRepo,
		log:          log,
	}
}

func (s *StatisticsService) GetStatistics(ctx context.Context) (*statistics.StatisticsResponse, error) {
	prs, err := s.prRepo.GetAllPRs(ctx)
	if err != nil {
		s.log.LogAttrs(ctx, slog.LevelError, "failed to get all PRs", slog.String("error", err.Error()))
		return nil, err
	}

	users, err := s.userRepo.GetAllUsers(ctx)
	if err != nil {
		s.log.LogAttrs(ctx, slog.LevelError, "failed to get all users", slog.String("error", err.Error()))
		return nil, err
	}

	reviewerCounts, err := s.reviewerRepo.GetAllReviewerCounts(ctx)
	if err != nil {
		s.log.LogAttrs(ctx, slog.LevelError, "failed to get reviewer counts", slog.String("error", err.Error()))
		return nil, err
	}

	totalPRs := len(prs)
	openPRs := 0
	mergedPRs := 0
	totalAssignments := 0

	prStats := make([]statistics.PRStats, 0, len(prs))
	for _, pr := range prs {
		if pr.Status == "OPEN" {
			openPRs++
		} else if pr.Status == "MERGED" {
			mergedPRs++
		}

		reviewers, err := s.reviewerRepo.GetReviewers(ctx, pr.Id)
		if err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to get reviewers for PR",
				slog.String("pr_id", pr.Id), slog.String("error", err.Error()))
			continue
		}

		totalAssignments += len(reviewers)

		prStats = append(prStats, statistics.PRStats{
			PullRequestID:   pr.Id,
			PullRequestName: pr.Title,
			ReviewersCount:  len(reviewers),
			Status:          pr.Status,
		})
	}

	userStatsMap := make(map[string]*statistics.UserStats)
	for _, user := range users {
		userStatsMap[user.Id] = &statistics.UserStats{
			UserID:           user.Id,
			Username:         user.Name,
			AssignmentsCount: reviewerCounts[user.Id],
			ActiveReviews:    0,
		}
	}

	for _, pr := range prs {
		if pr.Status != "OPEN" {
			continue
		}

		reviewers, err := s.reviewerRepo.GetReviewers(ctx, pr.Id)
		if err != nil {
			continue
		}

		for _, reviewerID := range reviewers {
			if stat, ok := userStatsMap[reviewerID]; ok {
				stat.ActiveReviews++
			}
		}
	}

	userStats := make([]statistics.UserStats, 0, len(userStatsMap))
	for _, stat := range userStatsMap {
		userStats = append(userStats, *stat)
	}

	s.log.LogAttrs(ctx, slog.LevelInfo, "statistics retrieved",
		slog.Int("total_prs", totalPRs),
		slog.Int("total_assignments", totalAssignments))

	return &statistics.StatisticsResponse{
		TotalPRs:         totalPRs,
		OpenPRs:          openPRs,
		MergedPRs:        mergedPRs,
		TotalAssignments: totalAssignments,
		UserStats:        userStats,
		PRStats:          prStats,
	}, nil
}
