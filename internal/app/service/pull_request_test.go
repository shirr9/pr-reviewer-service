package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/shirr9/pr-reviewer-service/internal/app/dto/pullrequest"
	"github.com/shirr9/pr-reviewer-service/internal/app/service/mocks"
	"github.com/shirr9/pr-reviewer-service/internal/domain/errors"
	"github.com/shirr9/pr-reviewer-service/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPullRequestService_CreatePR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPRRepo := mocks.NewMockPullRequestRepository(ctrl)
	mockReviewerRepo := mocks.NewMockReviewerRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUoW := mocks.NewMockTransactor(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewPullRequestService(mockPRRepo, mockReviewerRepo, mockUserRepo, mockUoW, logger)

	t.Run("Success - Create PR with 2 reviewers", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.CreatePrRequest{
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        "u1",
		}

		author := &models.User{
			Id:       "u1",
			Name:     "Alice",
			TeamName: "backend",
			IsActive: true,
		}

		candidates := []*models.User{
			{Id: "u2", Name: "Bob", TeamName: "backend", IsActive: true},
			{Id: "u3", Name: "Charlie", TeamName: "backend", IsActive: true},
		}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().Exists(ctx, "pr-1").Return(false, nil)
				mockUserRepo.EXPECT().FindByID(ctx, "u1").Return(author, nil)
				mockUserRepo.EXPECT().FindActiveCandidatesForReassignment(ctx, "backend", []string{"u1"}).Return(candidates, nil)
				mockPRRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
				mockReviewerRepo.EXPECT().AssignReviewer(ctx, "pr-1", "u2").Return(nil)
				mockReviewerRepo.EXPECT().AssignReviewer(ctx, "pr-1", "u3").Return(nil)
				return fn(ctx)
			},
		)

		resp, err := service.CreatePR(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "pr-1", resp.Pr.PullRequestID)
		assert.Equal(t, "Test PR", resp.Pr.PullRequestName)
		assert.Equal(t, "u1", resp.Pr.AuthorID)
		assert.Equal(t, models.PRStatusOpen, resp.Pr.Status)
		assert.Len(t, resp.Pr.AssignedReviewers, 2)
	})

	t.Run("Success - Create PR with 1 reviewer", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.CreatePrRequest{
			PullRequestID:   "pr-2",
			PullRequestName: "Another PR",
			AuthorID:        "u1",
		}

		author := &models.User{
			Id:       "u1",
			Name:     "Alice",
			TeamName: "backend",
			IsActive: true,
		}

		candidates := []*models.User{
			{Id: "u2", Name: "Bob", TeamName: "backend", IsActive: true},
		}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().Exists(ctx, "pr-2").Return(false, nil)
				mockUserRepo.EXPECT().FindByID(ctx, "u1").Return(author, nil)
				mockUserRepo.EXPECT().FindActiveCandidatesForReassignment(ctx, "backend", []string{"u1"}).Return(candidates, nil)
				mockPRRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
				mockReviewerRepo.EXPECT().AssignReviewer(ctx, "pr-2", "u2").Return(nil)
				return fn(ctx)
			},
		)

		resp, err := service.CreatePR(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Pr.AssignedReviewers, 1)
	})

	t.Run("Error - PR already exists", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.CreatePrRequest{
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        "u1",
		}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().Exists(ctx, "pr-1").Return(true, nil)
				return fn(ctx)
			},
		)

		resp, err := service.CreatePR(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "PR_EXISTS", err.(*errors.AppError).Code)
	})

	t.Run("Error - Author not found", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.CreatePrRequest{
			PullRequestID:   "pr-3",
			PullRequestName: "Test PR",
			AuthorID:        "nonexistent",
		}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().Exists(ctx, "pr-3").Return(false, nil)
				mockUserRepo.EXPECT().FindByID(ctx, "nonexistent").Return(nil, nil)
				return fn(ctx)
			},
		)

		resp, err := service.CreatePR(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "NOT_FOUND", err.(*errors.AppError).Code)
	})

	t.Run("Error - Author is not active", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.CreatePrRequest{
			PullRequestID:   "pr-4",
			PullRequestName: "Test PR",
			AuthorID:        "u1",
		}

		author := &models.User{
			Id:       "u1",
			Name:     "Alice",
			TeamName: "backend",
			IsActive: false,
		}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().Exists(ctx, "pr-4").Return(false, nil)
				mockUserRepo.EXPECT().FindByID(ctx, "u1").Return(author, nil)
				return fn(ctx)
			},
		)

		resp, err := service.CreatePR(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "NOT_FOUND", err.(*errors.AppError).Code)
	})

	t.Run("Success - Create PR with no reviewers", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.CreatePrRequest{
			PullRequestID:   "pr-5",
			PullRequestName: "Solo PR",
			AuthorID:        "u1",
		}

		author := &models.User{
			Id:       "u1",
			Name:     "Alice",
			TeamName: "backend",
			IsActive: true,
		}

		candidates := []*models.User{}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().Exists(ctx, "pr-5").Return(false, nil)
				mockUserRepo.EXPECT().FindByID(ctx, "u1").Return(author, nil)
				mockUserRepo.EXPECT().FindActiveCandidatesForReassignment(ctx, "backend", []string{"u1"}).Return(candidates, nil)
				mockPRRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
				return fn(ctx)
			},
		)

		resp, err := service.CreatePR(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Pr.AssignedReviewers, 0)
	})
}

func TestPullRequestService_MergePR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPRRepo := mocks.NewMockPullRequestRepository(ctrl)
	mockReviewerRepo := mocks.NewMockReviewerRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUoW := mocks.NewMockTransactor(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewPullRequestService(mockPRRepo, mockReviewerRepo, mockUserRepo, mockUoW, logger)

	t.Run("Success - Merge PR", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.MergePrRequest{
			PullRequestID: "pr-1",
		}

		pr := &models.PullRequest{
			Id:       "pr-1",
			Title:    "Test PR",
			AuthorId: "u1",
			Status:   models.PRStatusOpen,
		}

		reviewers := []string{"u2", "u3"}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().FindByID(ctx, "pr-1").Return(pr, nil)
				mockReviewerRepo.EXPECT().GetReviewers(ctx, "pr-1").Return(reviewers, nil)
				mockPRRepo.EXPECT().UpdateStatus(ctx, "pr-1", models.PRStatusMerged, gomock.Any()).Return(nil)
				return fn(ctx)
			},
		)

		resp, err := service.MergePR(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "pr-1", resp.Pr.PullRequestID)
		assert.Equal(t, models.PRStatusMerged, resp.Pr.Status)
		assert.NotEmpty(t, resp.Pr.MergedAt)
	})

	t.Run("Success - Idempotent merge (already merged)", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.MergePrRequest{
			PullRequestID: "pr-1",
		}

		mergedAt := time.Now().UTC()
		pr := &models.PullRequest{
			Id:       "pr-1",
			Title:    "Test PR",
			AuthorId: "u1",
			Status:   models.PRStatusMerged,
			MergedAt: &mergedAt,
		}

		reviewers := []string{"u2", "u3"}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().FindByID(ctx, "pr-1").Return(pr, nil)
				mockReviewerRepo.EXPECT().GetReviewers(ctx, "pr-1").Return(reviewers, nil)
				// UpdateStatus should NOT be called for idempotent case
				return fn(ctx)
			},
		)

		resp, err := service.MergePR(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "pr-1", resp.Pr.PullRequestID)
		assert.Equal(t, models.PRStatusMerged, resp.Pr.Status)
	})

	t.Run("Error - PR not found", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.MergePrRequest{
			PullRequestID: "nonexistent",
		}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().FindByID(ctx, "nonexistent").Return(nil, nil)
				return fn(ctx)
			},
		)

		resp, err := service.MergePR(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "NOT_FOUND", err.(*errors.AppError).Code)
	})
}

func TestPullRequestService_ReassignReviewer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPRRepo := mocks.NewMockPullRequestRepository(ctrl)
	mockReviewerRepo := mocks.NewMockReviewerRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUoW := mocks.NewMockTransactor(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewPullRequestService(mockPRRepo, mockReviewerRepo, mockUserRepo, mockUoW, logger)

	t.Run("Success - Reassign reviewer", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.ReassignReviewerRequest{
			PullRequestID: "pr-1",
			OldReviewerID: "u2",
		}

		pr := &models.PullRequest{
			Id:       "pr-1",
			Title:    "Test PR",
			AuthorId: "u1",
			Status:   models.PRStatusOpen,
		}

		oldReviewer := &models.User{
			Id:       "u2",
			Name:     "Bob",
			TeamName: "backend",
			IsActive: true,
		}

		currentReviewers := []string{"u2", "u3"}
		candidates := []*models.User{
			{Id: "u4", Name: "David", TeamName: "backend", IsActive: true},
		}
		updatedReviewers := []string{"u4", "u3"}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().FindByID(ctx, "pr-1").Return(pr, nil)
				mockReviewerRepo.EXPECT().IsAssigned(ctx, "pr-1", "u2").Return(true, nil)
				mockUserRepo.EXPECT().FindByID(ctx, "u2").Return(oldReviewer, nil)
				mockReviewerRepo.EXPECT().GetReviewers(ctx, "pr-1").Return(currentReviewers, nil)
				mockUserRepo.EXPECT().FindActiveCandidatesForReassignment(ctx, "backend", []string{"u1", "u2", "u3"}).Return(candidates, nil)
				mockReviewerRepo.EXPECT().ReplaceReviewer(ctx, "pr-1", "u2", "u4").Return(nil)
				mockReviewerRepo.EXPECT().GetReviewers(ctx, "pr-1").Return(updatedReviewers, nil)
				return fn(ctx)
			},
		)

		resp, err := service.ReassignReviewer(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "pr-1", resp.Pr.PullRequestID)
		assert.Equal(t, "u4", resp.ReplacedBy)
		assert.Equal(t, models.PRStatusOpen, resp.Pr.Status)
	})

	t.Run("Error - PR not found", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.ReassignReviewerRequest{
			PullRequestID: "nonexistent",
			OldReviewerID: "u2",
		}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().FindByID(ctx, "nonexistent").Return(nil, nil)
				return fn(ctx)
			},
		)

		resp, err := service.ReassignReviewer(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "NOT_FOUND", err.(*errors.AppError).Code)
	})

	t.Run("Error - PR is merged", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.ReassignReviewerRequest{
			PullRequestID: "pr-1",
			OldReviewerID: "u2",
		}

		pr := &models.PullRequest{
			Id:       "pr-1",
			Title:    "Test PR",
			AuthorId: "u1",
			Status:   models.PRStatusMerged,
		}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().FindByID(ctx, "pr-1").Return(pr, nil)
				return fn(ctx)
			},
		)

		resp, err := service.ReassignReviewer(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "PR_MERGED", err.(*errors.AppError).Code)
	})

	t.Run("Error - Reviewer not assigned", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.ReassignReviewerRequest{
			PullRequestID: "pr-1",
			OldReviewerID: "u5",
		}

		pr := &models.PullRequest{
			Id:       "pr-1",
			Title:    "Test PR",
			AuthorId: "u1",
			Status:   models.PRStatusOpen,
		}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().FindByID(ctx, "pr-1").Return(pr, nil)
				mockReviewerRepo.EXPECT().IsAssigned(ctx, "pr-1", "u5").Return(false, nil)
				return fn(ctx)
			},
		)

		resp, err := service.ReassignReviewer(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "NOT_ASSIGNED", err.(*errors.AppError).Code)
	})

	t.Run("Error - No candidates for reassignment", func(t *testing.T) {
		ctx := context.Background()
		req := pullrequest.ReassignReviewerRequest{
			PullRequestID: "pr-1",
			OldReviewerID: "u2",
		}

		pr := &models.PullRequest{
			Id:       "pr-1",
			Title:    "Test PR",
			AuthorId: "u1",
			Status:   models.PRStatusOpen,
		}

		oldReviewer := &models.User{
			Id:       "u2",
			Name:     "Bob",
			TeamName: "backend",
			IsActive: true,
		}

		currentReviewers := []string{"u2", "u3"}
		candidates := []*models.User{}

		mockUoW.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				mockPRRepo.EXPECT().FindByID(ctx, "pr-1").Return(pr, nil)
				mockReviewerRepo.EXPECT().IsAssigned(ctx, "pr-1", "u2").Return(true, nil)
				mockUserRepo.EXPECT().FindByID(ctx, "u2").Return(oldReviewer, nil)
				mockReviewerRepo.EXPECT().GetReviewers(ctx, "pr-1").Return(currentReviewers, nil)
				mockUserRepo.EXPECT().FindActiveCandidatesForReassignment(ctx, "backend", []string{"u1", "u2", "u3"}).Return(candidates, nil)
				return fn(ctx)
			},
		)

		resp, err := service.ReassignReviewer(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "NO_CANDIDATE", err.(*errors.AppError).Code)
	})
}
