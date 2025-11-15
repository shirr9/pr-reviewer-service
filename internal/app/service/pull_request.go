package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/shirr9/pr-reviewer-service/internal/app/dto/pullrequest"
	"github.com/shirr9/pr-reviewer-service/internal/domain/errors"
	"github.com/shirr9/pr-reviewer-service/internal/domain/models"
)

// PullRequestRepository defines the interface for pull request data persistence operations.
type PullRequestRepository interface {
	Create(ctx context.Context, pr *models.PullRequest) error
	FindByID(ctx context.Context, prID string) (*models.PullRequest, error)
	Exists(ctx context.Context, prID string) (bool, error)
	UpdateStatus(ctx context.Context, prID, status string, mergedAt *time.Time) error
}

// ReviewerRepository defines the interface for reviewer assignment operations.
type ReviewerRepository interface {
	AssignReviewer(ctx context.Context, prID, reviewerID string) error
	GetReviewers(ctx context.Context, prID string) ([]string, error)
	GetPRsByReviewer(ctx context.Context, reviewerID string) ([]string, error)
	IsAssigned(ctx context.Context, prID, reviewerID string) (bool, error)
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
}

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	FindByID(ctx context.Context, userID string) (*models.User, error)
	FindActiveCandidatesForReassignment(ctx context.Context, teamName string, excludeUserIDs []string) ([]*models.User, error)
}

// Transactor provides transaction management.
type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// PullRequestService implements business logic for managing pull requests.
type PullRequestService struct {
	prRepo       PullRequestRepository
	reviewerRepo ReviewerRepository
	userRepo     UserRepository
	uow          Transactor
	log          *slog.Logger
}

// NewPullRequestService creates a new pull request service.
func NewPullRequestService(
	prRepo PullRequestRepository,
	reviewerRepo ReviewerRepository,
	userRepo UserRepository,
	uow Transactor,
	log *slog.Logger,
) *PullRequestService {
	if log == nil {
		log = slog.Default()
	}
	return &PullRequestService{
		prRepo:       prRepo,
		reviewerRepo: reviewerRepo,
		userRepo:     userRepo,
		uow:          uow,
		log:          log,
	}
}

// CreatePR creates a new pull request and assigns up to 2 reviewers atomically.
// Uses Unit of Work pattern with Repeatable Read isolation level.
func (s *PullRequestService) CreatePR(ctx context.Context,
	req pullrequest.CreatePrRequest) (*pullrequest.CreatePrResponse, error) {

	var response pullrequest.CreatePrResponse
	var reviewerIDs []string

	err := s.uow.WithinTransaction(ctx, func(txCtx context.Context) error {
		exists, err := s.prRepo.Exists(txCtx, req.PullRequestID)
		if err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to check PR existence",
				slog.String("pr_id", req.PullRequestID), slog.String("error", err.Error()))
			return err
		}
		if exists {
			s.log.LogAttrs(ctx, slog.LevelWarn, "PR already exists",
				slog.String("pr_id", req.PullRequestID))
			return errors.NewPRExists("PR id already exists")
		}

		author, err := s.userRepo.FindByID(txCtx, req.AuthorID)
		if err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to find author",
				slog.String("author_id", req.AuthorID), slog.String("error", err.Error()))
			return err
		}
		if author == nil {
			s.log.LogAttrs(ctx, slog.LevelWarn, "author not found",
				slog.String("author_id", req.AuthorID))
			return errors.NewNotFound("resource not found")
		}

		if !author.IsActive {
			s.log.LogAttrs(ctx, slog.LevelWarn, "author is not active",
				slog.String("author_id", req.AuthorID))
			return errors.NewNotFound("author is not active")
		}

		if author.TeamName == "" {
			s.log.LogAttrs(ctx, slog.LevelWarn, "author has no team",
				slog.String("author_id", req.AuthorID))
			return errors.NewNotFound("resource not found")
		}

		candidates, err := s.userRepo.FindActiveCandidatesForReassignment(
			txCtx,
			author.TeamName,
			[]string{req.AuthorID})

		if err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to find reviewer candidates",
				slog.String("team", author.TeamName), slog.String("error", err.Error()))
			return err
		}
		const maxReviewers = 2
		reviewers := maxReviewers
		if len(candidates) < reviewers {
			reviewers = len(candidates)
		}

		if reviewers == 0 {
			s.log.LogAttrs(ctx, slog.LevelWarn, "no active reviewer candidates found",
				slog.String("pr_id", req.PullRequestID),
				slog.String("team", author.TeamName))
		}

		reviewerIDs = make([]string, 0, reviewers)
		for i := 0; i < reviewers; i++ {
			reviewerIDs = append(reviewerIDs, candidates[i].Id)
		}

		now := time.Now().UTC()
		pr := &models.PullRequest{
			Id:        req.PullRequestID,
			Title:     req.PullRequestName,
			AuthorId:  req.AuthorID,
			Status:    models.PRStatusOpen,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := s.prRepo.Create(txCtx, pr); err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to create PR",
				slog.String("pr_id", req.PullRequestID), slog.String("error", err.Error()))
			return err
		}

		for _, reviewerID := range reviewerIDs {
			if err := s.reviewerRepo.AssignReviewer(txCtx, req.PullRequestID, reviewerID); err != nil {
				s.log.LogAttrs(ctx, slog.LevelError, "failed to assign reviewer",
					slog.String("pr_id", req.PullRequestID),
					slog.String("reviewer_id", reviewerID),
					slog.String("error", err.Error()))
				return err
			}
		}
		response = pullrequest.CreatePrResponse{
			Pr: pullrequest.PR{
				PullRequestID:     pr.Id,
				PullRequestName:   pr.Title,
				AuthorID:          pr.AuthorId,
				Status:            pr.Status,
				AssignedReviewers: reviewerIDs,
			},
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.log.LogAttrs(ctx, slog.LevelInfo, "PR created successfully",
		slog.String("pr_id", req.PullRequestID),
		slog.Int("reviewers_count", len(reviewerIDs)))
	return &response, nil
}

// MergePR marks PR as MERGED (idempotent operation).
func (s *PullRequestService) MergePR(ctx context.Context, req pullrequest.MergePrRequest) (*pullrequest.MergePrResponse, error) {
	var response pullrequest.MergePrResponse

	err := s.uow.WithinTransaction(ctx, func(txCtx context.Context) error {
		pr, err := s.prRepo.FindByID(txCtx, req.PullRequestID)
		if err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to find PR",
				slog.String("pr_id", req.PullRequestID), slog.String("error", err.Error()))
			return err
		}
		if pr == nil {
			s.log.LogAttrs(ctx, slog.LevelWarn, "PR not found",
				slog.String("pr_id", req.PullRequestID))
			return errors.NewNotFound("PR not found")
		}

		reviewers, err := s.reviewerRepo.GetReviewers(txCtx, pr.Id)
		if err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to get reviewers",
				slog.String("pr_id", pr.Id), slog.String("error", err.Error()))
			return err
		}

		if pr.Status == models.PRStatusMerged {
			s.log.LogAttrs(ctx, slog.LevelInfo, "PR already merged (idempotent)",
				slog.String("pr_id", pr.Id))

			response = pullrequest.MergePrResponse{
				Pr: pullrequest.PR{
					PullRequestID:     pr.Id,
					PullRequestName:   pr.Title,
					AuthorID:          pr.AuthorId,
					Status:            pr.Status,
					AssignedReviewers: reviewers,
					MergedAt:          pr.MergedAt.Format(time.RFC3339),
				},
			}
			return nil
		}

		mergedAt := time.Now().UTC()
		if err := s.prRepo.UpdateStatus(txCtx, pr.Id, models.PRStatusMerged, &mergedAt); err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to update PR status",
				slog.String("pr_id", pr.Id), slog.String("error", err.Error()))
			return err
		}

		response = pullrequest.MergePrResponse{
			Pr: pullrequest.PR{
				PullRequestID:     pr.Id,
				PullRequestName:   pr.Title,
				AuthorID:          pr.AuthorId,
				Status:            models.PRStatusMerged,
				AssignedReviewers: reviewers,
				MergedAt:          mergedAt.Format(time.RFC3339),
			},
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.log.LogAttrs(ctx, slog.LevelInfo, "PR merged successfully",
		slog.String("pr_id", req.PullRequestID))

	return &response, nil
}

// ReassignReviewer replaces old reviewer with a new one from the same team.
func (s *PullRequestService) ReassignReviewer(ctx context.Context, req pullrequest.ReassignReviewerRequest) (*pullrequest.ReassignReviewerResponse, error) {
	var response pullrequest.ReassignReviewerResponse

	err := s.uow.WithinTransaction(ctx, func(txCtx context.Context) error {
		pr, err := s.prRepo.FindByID(txCtx, req.PullRequestID)
		if err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to find PR",
				slog.String("pr_id", req.PullRequestID), slog.String("error", err.Error()))
			return err
		}
		if pr == nil {
			s.log.LogAttrs(ctx, slog.LevelWarn, "PR not found",
				slog.String("pr_id", req.PullRequestID))
			return errors.NewNotFound("PR not found")
		}

		if pr.Status == models.PRStatusMerged {
			s.log.LogAttrs(ctx, slog.LevelWarn, "cannot reassign on merged PR",
				slog.String("pr_id", req.PullRequestID))
			return errors.NewPRMerged("cannot reassign on merged PR")
		}

		isAssigned, err := s.reviewerRepo.IsAssigned(txCtx, req.PullRequestID, req.OldReviewerID)
		if err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to check reviewer assignment",
				slog.String("pr_id", req.PullRequestID),
				slog.String("reviewer_id", req.OldReviewerID),
				slog.String("error", err.Error()))
			return err
		}
		if !isAssigned {
			s.log.LogAttrs(ctx, slog.LevelWarn, "reviewer is not assigned to this PR",
				slog.String("pr_id", req.PullRequestID),
				slog.String("reviewer_id", req.OldReviewerID))
			return errors.NewNotAssigned("reviewer is not assigned to this PR")
		}

		oldReviewer, err := s.userRepo.FindByID(txCtx, req.OldReviewerID)
		if err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to find old reviewer",
				slog.String("reviewer_id", req.OldReviewerID), slog.String("error", err.Error()))
			return err
		}
		if oldReviewer == nil {
			s.log.LogAttrs(ctx, slog.LevelWarn, "old reviewer not found",
				slog.String("reviewer_id", req.OldReviewerID))
			return errors.NewNotFound("old reviewer not found")
		}

		currentReviewers, err := s.reviewerRepo.GetReviewers(txCtx, req.PullRequestID)
		if err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to get current reviewers",
				slog.String("pr_id", req.PullRequestID), slog.String("error", err.Error()))
			return err
		}

		excludeUserIDs := append([]string{pr.AuthorId}, currentReviewers...)

		candidates, err := s.userRepo.FindActiveCandidatesForReassignment(
			txCtx,
			oldReviewer.TeamName,
			excludeUserIDs,
		)
		if err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to find replacement candidates",
				slog.String("team", oldReviewer.TeamName), slog.String("error", err.Error()))
			return err
		}

		if len(candidates) == 0 {
			s.log.LogAttrs(ctx, slog.LevelWarn, "no active replacement candidate in team",
				slog.String("team", oldReviewer.TeamName))
			return errors.NewNoCandidate("no active replacement candidate in team")
		}

		newReviewerID := candidates[0].Id

		if err := s.reviewerRepo.ReplaceReviewer(txCtx, req.PullRequestID, req.OldReviewerID, newReviewerID); err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to replace reviewer",
				slog.String("pr_id", req.PullRequestID),
				slog.String("old_reviewer", req.OldReviewerID),
				slog.String("new_reviewer", newReviewerID),
				slog.String("error", err.Error()))
			return err
		}

		updatedReviewers, err := s.reviewerRepo.GetReviewers(txCtx, req.PullRequestID)
		if err != nil {
			s.log.LogAttrs(ctx, slog.LevelError, "failed to get updated reviewers",
				slog.String("pr_id", req.PullRequestID), slog.String("error", err.Error()))
			return err
		}

		response = pullrequest.ReassignReviewerResponse{
			Pr: pullrequest.PR{
				PullRequestID:     pr.Id,
				PullRequestName:   pr.Title,
				AuthorID:          pr.AuthorId,
				Status:            pr.Status,
				AssignedReviewers: updatedReviewers,
			},
			ReplacedBy: newReviewerID,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.log.LogAttrs(ctx, slog.LevelInfo, "reviewer reassigned successfully",
		slog.String("pr_id", req.PullRequestID),
		slog.String("old_reviewer", req.OldReviewerID))
	return &response, nil
}
