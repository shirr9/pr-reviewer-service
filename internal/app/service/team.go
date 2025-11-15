package service

import (
	"context"
	"log/slog"

	"github.com/shirr9/pr-reviewer-service/internal/app/dto/team"
	"github.com/shirr9/pr-reviewer-service/internal/domain/errors"
	"github.com/shirr9/pr-reviewer-service/internal/domain/models"
)

// TeamRepository defines the interface for team and user management operations.
type TeamRepository interface {
	CreateOrUpdateTeam(ctx context.Context, team *models.Team) error
	GetTeamByName(ctx context.Context, teamName string) (*models.Team, error)
	IsExists(ctx context.Context, teamName string) (bool, error)
}

// TeamService implements business logic for managing teams.
type TeamService struct {
	teamRepo TeamRepository
	log      *slog.Logger
}

// NewTeamService creates a new team service.
func NewTeamService(teamRepo TeamRepository, log *slog.Logger) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		log:      log,
	}
}

// AddTeam creates a new team with members (creates/updates users).
func (s *TeamService) AddTeam(ctx context.Context, req team.AddTeamRequest) (team.AddTeamResponse, error) {
	if len(req.Members) == 0 {
		s.log.LogAttrs(ctx, slog.LevelWarn, "team must have at least one member",
			slog.String("team_name", req.TeamName))
		return team.AddTeamResponse{}, errors.New("BAD_REQUEST", "team must have at least one member")
	}

	exists, err := s.teamRepo.IsExists(ctx, req.TeamName)
	if err != nil {
		s.log.LogAttrs(ctx, slog.LevelError, "failed to check team existence",
			slog.String("team_name", req.TeamName), slog.String("error", err.Error()))
		return team.AddTeamResponse{}, err
	}

	if exists {
		s.log.LogAttrs(ctx, slog.LevelWarn, "team already exists",
			slog.String("team_name", req.TeamName))
		return team.AddTeamResponse{}, errors.NewTeamExists("team_name already exists")
	}

	domainTeam := &models.Team{
		Members: make([]*models.User, 0, len(req.Members)),
	}

	for _, memberDTO := range req.Members {
		domainTeam.Members = append(domainTeam.Members, &models.User{
			Id:       memberDTO.UserID,
			Name:     memberDTO.Username,
			TeamName: req.TeamName,
			IsActive: memberDTO.IsActive,
		})
	}

	if err := s.teamRepo.CreateOrUpdateTeam(ctx, domainTeam); err != nil {
		s.log.LogAttrs(ctx, slog.LevelError, "failed to create team",
			slog.String("team_name", req.TeamName), slog.String("error", err.Error()))
		return team.AddTeamResponse{}, err
	}

	s.log.LogAttrs(ctx, slog.LevelInfo, "team created successfully",
		slog.String("team_name", req.TeamName),
		slog.Int("members_count", len(req.Members)))

	return team.AddTeamResponse{
		Team: team.Team{
			TeamName: req.TeamName,
			Members:  req.Members,
		},
	}, nil
}

// GetTeam returns a team with all its members.
func (s *TeamService) GetTeam(ctx context.Context, teamName string) (team.GetTeamResponse, error) {
	t, err := s.teamRepo.GetTeamByName(ctx, teamName)
	if err != nil {
		s.log.LogAttrs(ctx, slog.LevelError, "failed to get team",
			slog.String("team_name", teamName), slog.String("error", err.Error()))
		return team.GetTeamResponse{}, err
	}

	if t == nil {
		s.log.LogAttrs(ctx, slog.LevelWarn, "team not found",
			slog.String("team_name", teamName))
		return team.GetTeamResponse{}, errors.NewNotFound("team not found")
	}

	members := make([]team.TeamMember, 0, len(t.Members))
	for _, user := range t.Members {
		members = append(members, team.TeamMember{
			UserID:   user.Id,
			Username: user.Name,
			IsActive: user.IsActive,
		})
	}

	s.log.LogAttrs(ctx, slog.LevelInfo, "team retrieved successfully",
		slog.String("team_name", teamName),
		slog.Int("members_count", len(members)))

	return team.GetTeamResponse{
		TeamName: teamName,
		Members:  members,
	}, nil
}
