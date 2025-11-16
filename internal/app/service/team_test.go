package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/shirr9/pr-reviewer-service/internal/app/dto/team"
	"github.com/shirr9/pr-reviewer-service/internal/app/service/mocks"
	"github.com/shirr9/pr-reviewer-service/internal/domain/errors"
	"github.com/shirr9/pr-reviewer-service/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTeamService_AddTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamRepo := mocks.NewMockTeamRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewTeamService(mockTeamRepo, nil, nil, nil, nil, logger)

	t.Run("Success - Add new team", func(t *testing.T) {
		ctx := context.Background()
		req := team.AddTeamRequest{
			TeamName: "backend",
			Members: []team.TeamMember{
				{UserID: "u1", Username: "Alice", IsActive: true},
				{UserID: "u2", Username: "Bob", IsActive: true},
				{UserID: "u3", Username: "Charlie", IsActive: false},
			},
		}

		mockTeamRepo.EXPECT().IsExists(ctx, "backend").Return(false, nil)
		mockTeamRepo.EXPECT().CreateOrUpdateTeam(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, team *models.Team) error {
				assert.Len(t, team.Members, 3)
				assert.Equal(t, "u1", team.Members[0].Id)
				assert.Equal(t, "Alice", team.Members[0].Name)
				assert.Equal(t, "backend", team.Members[0].TeamName)
				assert.True(t, team.Members[0].IsActive)
				return nil
			},
		)

		resp, err := service.AddTeam(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "backend", resp.Team.TeamName)
		assert.Len(t, resp.Team.Members, 3)
	})

	t.Run("Error - Team already exists", func(t *testing.T) {
		ctx := context.Background()
		req := team.AddTeamRequest{
			TeamName: "backend",
			Members: []team.TeamMember{
				{UserID: "u1", Username: "Alice", IsActive: true},
			},
		}

		mockTeamRepo.EXPECT().IsExists(ctx, "backend").Return(true, nil)

		resp, err := service.AddTeam(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "TEAM_EXISTS", err.(*errors.AppError).Code)
	})

	t.Run("Error - Empty members list", func(t *testing.T) {
		ctx := context.Background()
		req := team.AddTeamRequest{
			TeamName: "emptyteam",
			Members:  []team.TeamMember{},
		}

		resp, err := service.AddTeam(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "BAD_REQUEST", err.(*errors.AppError).Code)
		assert.Contains(t, err.Error(), "at least one member")
	})

	t.Run("Success - Add team with single member", func(t *testing.T) {
		ctx := context.Background()
		req := team.AddTeamRequest{
			TeamName: "payments",
			Members: []team.TeamMember{
				{UserID: "u4", Username: "David", IsActive: true},
			},
		}

		mockTeamRepo.EXPECT().IsExists(ctx, "payments").Return(false, nil)
		mockTeamRepo.EXPECT().CreateOrUpdateTeam(ctx, gomock.Any()).Return(nil)

		resp, err := service.AddTeam(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "payments", resp.Team.TeamName)
		assert.Len(t, resp.Team.Members, 1)
	})

	t.Run("Success - Add team with inactive members", func(t *testing.T) {
		ctx := context.Background()
		req := team.AddTeamRequest{
			TeamName: "testteam",
			Members: []team.TeamMember{
				{UserID: "u5", Username: "Eve", IsActive: false},
				{UserID: "u6", Username: "Frank", IsActive: false},
			},
		}

		mockTeamRepo.EXPECT().IsExists(ctx, "testteam").Return(false, nil)
		mockTeamRepo.EXPECT().CreateOrUpdateTeam(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, team *models.Team) error {
				assert.False(t, team.Members[0].IsActive)
				assert.False(t, team.Members[1].IsActive)
				return nil
			},
		)

		resp, err := service.AddTeam(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestTeamService_GetTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamRepo := mocks.NewMockTeamRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewTeamService(mockTeamRepo, nil, nil, nil, nil, logger)

	t.Run("Success - Get existing team", func(t *testing.T) {
		ctx := context.Background()
		teamName := "backend"

		domainTeam := &models.Team{
			Members: []*models.User{
				{Id: "u1", Name: "Alice", TeamName: "backend", IsActive: true},
				{Id: "u2", Name: "Bob", TeamName: "backend", IsActive: true},
				{Id: "u3", Name: "Charlie", TeamName: "backend", IsActive: false},
			},
		}

		mockTeamRepo.EXPECT().GetTeamByName(ctx, "backend").Return(domainTeam, nil)

		resp, err := service.GetTeam(ctx, teamName)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "backend", resp.TeamName)
		assert.Len(t, resp.Members, 3)
		assert.Equal(t, "u1", resp.Members[0].UserID)
		assert.Equal(t, "Alice", resp.Members[0].Username)
		assert.True(t, resp.Members[0].IsActive)
		assert.Equal(t, "u3", resp.Members[2].UserID)
		assert.False(t, resp.Members[2].IsActive)
	})

	t.Run("Error - Team not found", func(t *testing.T) {
		ctx := context.Background()
		teamName := "nonexistent"

		mockTeamRepo.EXPECT().GetTeamByName(ctx, "nonexistent").Return(nil, nil)

		resp, err := service.GetTeam(ctx, teamName)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "NOT_FOUND", err.(*errors.AppError).Code)
	})

	t.Run("Success - Get team with single member", func(t *testing.T) {
		ctx := context.Background()
		teamName := "solo"

		domainTeam := &models.Team{
			Members: []*models.User{
				{Id: "u1", Name: "Solo", TeamName: "solo", IsActive: true},
			},
		}

		mockTeamRepo.EXPECT().GetTeamByName(ctx, "solo").Return(domainTeam, nil)

		resp, err := service.GetTeam(ctx, teamName)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "solo", resp.TeamName)
		assert.Len(t, resp.Members, 1)
	})

	t.Run("Success - Get team with all inactive members", func(t *testing.T) {
		ctx := context.Background()
		teamName := "inactive-team"

		domainTeam := &models.Team{
			Members: []*models.User{
				{Id: "u4", Name: "David", TeamName: "inactive-team", IsActive: false},
				{Id: "u5", Name: "Eve", TeamName: "inactive-team", IsActive: false},
			},
		}

		mockTeamRepo.EXPECT().GetTeamByName(ctx, "inactive-team").Return(domainTeam, nil)

		resp, err := service.GetTeam(ctx, teamName)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Members, 2)
		assert.False(t, resp.Members[0].IsActive)
		assert.False(t, resp.Members[1].IsActive)
	})
}
