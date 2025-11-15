package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shirr9/pr-reviewer-service/internal/domain/models"
)

// TeamRepository manages teams in the database.
type TeamRepository struct {
	pool *pgxpool.Pool
}

// CreateOrUpdateTeam creates/updates a team and its members.
func (r *TeamRepository) CreateOrUpdateTeam(ctx context.Context, team *models.Team) error {
	teamName := team.GetTeamName()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	upsertQuery := `
		INSERT INTO "user" (id, username, team_name, is_active) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) 
		DO UPDATE SET 
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active`

	for _, member := range team.Members {
		_, err = tx.Exec(ctx, upsertQuery,
			member.Id, member.Name, teamName, member.IsActive)
		if err != nil {
			return fmt.Errorf("failed to upsert user %s: %w", member.Id, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// IsExists checks if a team exists by team name.
func (r *TeamRepository) IsExists(ctx context.Context, teamName string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM "user" WHERE team_name = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}

	return exists, nil
}

// GetTeamByName gets a team by its name.
func (r *TeamRepository) GetTeamByName(ctx context.Context, teamName string) (*models.Team, error) {
	query := `
		SELECT id, username, team_name, is_active 
		FROM "user" 
		WHERE team_name = $1 
		ORDER BY username`

	rows, err := r.pool.Query(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	defer rows.Close()

	var members []*models.User
	for rows.Next() {
		var user models.User
		if err = rows.Scan(&user.Id, &user.Name, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		members = append(members, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	if len(members) == 0 {
		return nil, nil
	}

	return &models.Team{
		Members: members,
	}, nil
}
