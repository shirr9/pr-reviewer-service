package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shirr9/pr-reviewer-service/internal/domain/models"
)

type PullRequestRepository struct {
	pool *pgxpool.Pool
}

// Create creates a new Pull Request.
func (r *PullRequestRepository) Create(ctx context.Context, pr *models.PullRequest) error {
	query := `INSERT INTO pull_request (id, title, author_id, status, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6)`

	executor := getTx(ctx, r.pool)
	_, err := executor.Exec(ctx, query,
		pr.Id, pr.Title, pr.AuthorId, pr.Status, pr.CreatedAt, pr.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	return nil
}

// FindByID finds PR by ID.
func (r *PullRequestRepository) FindByID(ctx context.Context, prID string) (*models.PullRequest, error) {
	query := `SELECT id, title, author_id, status, created_at, merged_at, updated_at 
	          FROM pull_request 
	          WHERE id = $1`

	executor := getTx(ctx, r.pool)
	var pr models.PullRequest
	err := executor.QueryRow(ctx, query, prID).Scan(
		&pr.Id, &pr.Title, &pr.AuthorId, &pr.Status,
		&pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find pull request: %w", err)
	}

	return &pr, nil
}

// Exists checks if a PR exists by ID.
func (r *PullRequestRepository) Exists(ctx context.Context, prID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pull_request WHERE id = $1)`

	executor := getTx(ctx, r.pool)
	var exists bool
	err := executor.QueryRow(ctx, query, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check PR existence: %w", err)
	}

	return exists, nil
}

// UpdateStatus updates PR status.
func (r *PullRequestRepository) UpdateStatus(ctx context.Context, prID, status string, mergedAt *time.Time) error {
	query := `UPDATE pull_request 
	          SET status = $2, merged_at = $3, updated_at = $4 
	          WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, prID, status, mergedAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update PR status: %w", err)
	}

	return nil
}

// FindByReviewer finds all PR, where the user is assigned as a reviewer.
func (r *PullRequestRepository) FindByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequest, error) {
	query := `SELECT DISTINCT pr.id, pr.title, pr.author_id, pr.status, 
	                 pr.created_at, pr.merged_at, pr.updated_at
	          FROM pull_request pr
	          JOIN pr_reviewer prr ON pr.id = prr.pr_id
	          WHERE prr.reviewer_id = $1
	          ORDER BY pr.created_at DESC`

	rows, err := r.pool.Query(ctx, query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find PRs by reviewer: %w", err)
	}
	defer rows.Close()

	var prs []*models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		if err = rows.Scan(
			&pr.Id, &pr.Title, &pr.AuthorId, &pr.Status,
			&pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		prs = append(prs, &pr)
	}

	return prs, nil
}

// GetAllPRs returns all pull requests.
func (r *PullRequestRepository) GetAllPRs(ctx context.Context) ([]*models.PullRequest, error) {
	query := `SELECT id, title, author_id, status, created_at, merged_at, updated_at
	          FROM pull_request
	          ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all PRs: %w", err)
	}
	defer rows.Close()

	var prs []*models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		if err = rows.Scan(
			&pr.Id, &pr.Title, &pr.AuthorId, &pr.Status,
			&pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		prs = append(prs, &pr)
	}

	return prs, nil
}

// FindOpenPRsByReviewers finds all open PRs where any of the specified reviewers is assigned.
func (r *PullRequestRepository) FindOpenPRsByReviewers(ctx context.Context, reviewerIDs []string) ([]*models.PullRequest, error) {
	query := `SELECT DISTINCT pr.id, pr.title, pr.author_id, pr.status, 
	                 pr.created_at, pr.merged_at, pr.updated_at
	          FROM pull_request pr
	          JOIN pr_reviewer prr ON pr.id = prr.pr_id
	          WHERE prr.reviewer_id = ANY($1) AND pr.status = 'OPEN'
	          ORDER BY pr.created_at DESC`

	executor := getTx(ctx, r.pool)
	rows, err := executor.Query(ctx, query, reviewerIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find open PRs by reviewers: %w", err)
	}
	defer rows.Close()

	var prs []*models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		if err = rows.Scan(
			&pr.Id, &pr.Title, &pr.AuthorId, &pr.Status,
			&pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		prs = append(prs, &pr)
	}

	return prs, nil
}
