package service

import (
	"context"
	"log/slog"

	userDto "github.com/shirr9/pr-reviewer-service/internal/app/dto/user"
	"github.com/shirr9/pr-reviewer-service/internal/domain/errors"
	"github.com/shirr9/pr-reviewer-service/internal/domain/models"
)

// UserRepositoryForService defines the interface for user data operations needed by UserService.
type UserRepositoryForService interface {
	FindByID(ctx context.Context, userID string) (*models.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) error
}

// PullRequestRepositoryForUser defines the interface for PR operations needed by UserService.
type PullRequestRepositoryForUser interface {
	FindByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequest, error)
}

// UserService implements business logic for user operations.
type UserService struct {
	userRepo UserRepositoryForService
	prRepo   PullRequestRepositoryForUser
	log      *slog.Logger
}

// NewUserService creates a new user service.
func NewUserService(
	userRepo UserRepositoryForService,
	prRepo PullRequestRepositoryForUser,
	log *slog.Logger,
) *UserService {
	if log == nil {
		log = slog.Default()
	}
	return &UserService{
		userRepo: userRepo,
		prRepo:   prRepo,
		log:      log,
	}
}

// SetIsActive updates user's active status and returns updated user.
func (s *UserService) SetIsActive(ctx context.Context, req userDto.SetIsActiveRequest) (*userDto.SetIsActiveResponse, error) {
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		s.log.LogAttrs(ctx, slog.LevelError, "failed to find user",
			slog.String("user_id", req.UserID), slog.String("error", err.Error()))
		return nil, err
	}
	if user == nil {
		s.log.LogAttrs(ctx, slog.LevelWarn, "user not found",
			slog.String("user_id", req.UserID))
		return nil, errors.NewNotFound("user not found")
	}

	if err := s.userRepo.SetIsActive(ctx, req.UserID, req.IsActive); err != nil {
		s.log.LogAttrs(ctx, slog.LevelError, "failed to set is_active",
			slog.String("user_id", req.UserID),
			slog.Bool("is_active", req.IsActive),
			slog.String("error", err.Error()))
		return nil, err
	}

	s.log.LogAttrs(ctx, slog.LevelInfo, "user is_active updated",
		slog.String("user_id", req.UserID),
		slog.Bool("is_active", req.IsActive))

	return &userDto.SetIsActiveResponse{
		User: userDto.User{
			UserID:   user.Id,
			Username: user.Name,
			TeamName: user.TeamName,
			IsActive: req.IsActive,
		},
	}, nil
}

// GetReview returns list of PRs where user is assigned as reviewer.
func (s *UserService) GetReview(ctx context.Context, userID string) (*userDto.GetReviewResponse, error) {
	prs, err := s.prRepo.FindByReviewer(ctx, userID)
	if err != nil {
		s.log.LogAttrs(ctx, slog.LevelError, "failed to find PRs by reviewer",
			slog.String("user_id", userID), slog.String("error", err.Error()))
		return nil, err
	}

	prDTOs := make([]userDto.PR, 0, len(prs))
	for _, pr := range prs {
		prDTOs = append(prDTOs, userDto.PR{
			PullRequestID:   pr.Id,
			PullRequestName: pr.Title,
			AuthorID:        pr.AuthorId,
			Status:          pr.Status,
		})
	}

	s.log.LogAttrs(ctx, slog.LevelInfo, "user PRs retrieved",
		slog.String("user_id", userID),
		slog.Int("pr_count", len(prDTOs)))

	return &userDto.GetReviewResponse{
		UserID:       userID,
		PullRequests: prDTOs,
	}, nil
}
