package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/shirr9/pr-reviewer-service/internal/app/dto/user"
	"github.com/shirr9/pr-reviewer-service/internal/app/service/mocks"
	"github.com/shirr9/pr-reviewer-service/internal/domain/errors"
	"github.com/shirr9/pr-reviewer-service/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserService_SetIsActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryForService(ctrl)
	mockPRRepo := mocks.NewMockPullRequestRepositoryForUser(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewUserService(mockUserRepo, mockPRRepo, logger)

	t.Run("Success - Set user active", func(t *testing.T) {
		ctx := context.Background()
		req := user.SetIsActiveRequest{
			UserID:   "u1",
			IsActive: true,
		}

		existingUser := &models.User{
			Id:       "u1",
			Name:     "Alice",
			TeamName: "backend",
			IsActive: false,
		}

		mockUserRepo.EXPECT().FindByID(ctx, "u1").Return(existingUser, nil)
		mockUserRepo.EXPECT().SetIsActive(ctx, "u1", true).Return(nil)

		resp, err := service.SetIsActive(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "u1", resp.User.UserID)
		assert.Equal(t, "Alice", resp.User.Username)
		assert.True(t, resp.User.IsActive)
	})

	t.Run("Success - Set user inactive", func(t *testing.T) {
		ctx := context.Background()
		req := user.SetIsActiveRequest{
			UserID:   "u2",
			IsActive: false,
		}

		existingUser := &models.User{
			Id:       "u2",
			Name:     "Bob",
			TeamName: "backend",
			IsActive: true,
		}

		mockUserRepo.EXPECT().FindByID(ctx, "u2").Return(existingUser, nil)
		mockUserRepo.EXPECT().SetIsActive(ctx, "u2", false).Return(nil)

		resp, err := service.SetIsActive(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "u2", resp.User.UserID)
		assert.False(t, resp.User.IsActive)
	})

	t.Run("Error - User not found", func(t *testing.T) {
		ctx := context.Background()
		req := user.SetIsActiveRequest{
			UserID:   "nonexistent",
			IsActive: true,
		}

		mockUserRepo.EXPECT().FindByID(ctx, "nonexistent").Return(nil, nil)

		resp, err := service.SetIsActive(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "NOT_FOUND", err.(*errors.AppError).Code)
	})

	t.Run("Success - Idempotent operation (same state)", func(t *testing.T) {
		ctx := context.Background()
		req := user.SetIsActiveRequest{
			UserID:   "u3",
			IsActive: true,
		}

		existingUser := &models.User{
			Id:       "u3",
			Name:     "Charlie",
			TeamName: "backend",
			IsActive: true,
		}

		mockUserRepo.EXPECT().FindByID(ctx, "u3").Return(existingUser, nil)
		mockUserRepo.EXPECT().SetIsActive(ctx, "u3", true).Return(nil)

		resp, err := service.SetIsActive(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.User.IsActive)
	})
}

func TestUserService_GetReview(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepositoryForService(ctrl)
	mockPRRepo := mocks.NewMockPullRequestRepositoryForUser(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewUserService(mockUserRepo, mockPRRepo, logger)

	t.Run("Success - Get reviews for user with multiple PRs", func(t *testing.T) {
		ctx := context.Background()
		userID := "u2"

		prs := []*models.PullRequest{
			{
				Id:       "pr-1",
				Title:    "Add feature",
				AuthorId: "u1",
				Status:   models.PRStatusOpen,
			},
			{
				Id:       "pr-2",
				Title:    "Fix bug",
				AuthorId: "u3",
				Status:   models.PRStatusOpen,
			},
		}

		mockPRRepo.EXPECT().FindByReviewer(ctx, "u2").Return(prs, nil)

		resp, err := service.GetReview(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.PullRequests, 2)
		assert.Equal(t, "pr-1", resp.PullRequests[0].PullRequestID)
		assert.Equal(t, "Add feature", resp.PullRequests[0].PullRequestName)
		assert.Equal(t, "u1", resp.PullRequests[0].AuthorID)
		assert.Equal(t, models.PRStatusOpen, resp.PullRequests[0].Status)
	})

	t.Run("Success - Get reviews for user with no PRs", func(t *testing.T) {
		ctx := context.Background()
		userID := "u3"

		prs := []*models.PullRequest{}

		mockPRRepo.EXPECT().FindByReviewer(ctx, "u3").Return(prs, nil)

		resp, err := service.GetReview(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.PullRequests, 0)
	})

	t.Run("Success - Get reviews including merged PRs", func(t *testing.T) {
		ctx := context.Background()
		userID := "u1"

		prs := []*models.PullRequest{
			{
				Id:       "pr-1",
				Title:    "Open PR",
				AuthorId: "u2",
				Status:   models.PRStatusOpen,
			},
			{
				Id:       "pr-2",
				Title:    "Merged PR",
				AuthorId: "u3",
				Status:   models.PRStatusMerged,
			},
		}

		mockPRRepo.EXPECT().FindByReviewer(ctx, "u1").Return(prs, nil)

		resp, err := service.GetReview(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.PullRequests, 2)
		assert.Equal(t, models.PRStatusOpen, resp.PullRequests[0].Status)
		assert.Equal(t, models.PRStatusMerged, resp.PullRequests[1].Status)
	})

	t.Run("Success - Get reviews for inactive user", func(t *testing.T) {
		ctx := context.Background()
		userID := "u4"

		prs := []*models.PullRequest{
			{
				Id:       "pr-10",
				Title:    "Old PR",
				AuthorId: "u1",
				Status:   models.PRStatusOpen,
			},
		}

		mockPRRepo.EXPECT().FindByReviewer(ctx, "u4").Return(prs, nil)

		resp, err := service.GetReview(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.PullRequests, 1)
	})
}
