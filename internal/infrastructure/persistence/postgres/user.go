package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shirr9/pr-reviewer-service/internal/domain/models"
)

// UserRepository manages users in the database.
type UserRepository struct {
	pool *pgxpool.Pool
}

// FindByID finds user by ID.
func (r *UserRepository) FindByID(ctx context.Context, userID string) (*models.User, error) {
	query := `SELECT id, username, team_name, is_active FROM "user" WHERE id = $1`

	executor := getTx(ctx, r.pool)
	var user models.User
	err := executor.QueryRow(ctx, query, userID).Scan(
		&user.Id, &user.Name, &user.TeamName, &user.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// SetIsActive updates the is_active status of a user.
func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	query := `UPDATE "user" SET is_active = $2 WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, userID, isActive)
	if err != nil {
		return fmt.Errorf("failed to set is_active: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil
	}

	return nil
}

// FindActiveCandidatesForReassignment finds active users in the same team excluding specified user IDs.
func (r *UserRepository) FindActiveCandidatesForReassignment(ctx context.Context, teamName string, excludeUserIDs []string) ([]*models.User, error) {
	query := `SELECT id, username, team_name, is_active 
	          FROM "user" 
	          WHERE team_name = $1 AND is_active = true AND id != ALL($2)
	          ORDER BY id`

	executor := getTx(ctx, r.pool)
	rows, err := executor.Query(ctx, query, teamName, excludeUserIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find active candidates: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.Id, &user.Name, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// GetAllUsers returns all users.
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	query := `SELECT id, username, team_name, is_active FROM "user" ORDER BY id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.Id, &user.Name, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// FindByTeamName finds all users in a team.
func (r *UserRepository) FindByTeamName(ctx context.Context, teamName string) ([]*models.User, error) {
	query := `SELECT id, username, team_name, is_active 
	          FROM "user" 
	          WHERE team_name = $1
	          ORDER BY id`

	executor := getTx(ctx, r.pool)
	rows, err := executor.Query(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to find users by team: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.Id, &user.Name, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// DeactivateTeamUsers deactivates all users in a team.
func (r *UserRepository) DeactivateTeamUsers(ctx context.Context, teamName string) (int, error) {
	query := `UPDATE "user" SET is_active = false WHERE team_name = $1 AND is_active = true`

	executor := getTx(ctx, r.pool)
	result, err := executor.Exec(ctx, query, teamName)
	if err != nil {
		return 0, fmt.Errorf("failed to deactivate team users: %w", err)
	}

	return int(result.RowsAffected()), nil
}
