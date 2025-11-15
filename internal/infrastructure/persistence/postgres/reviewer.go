package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ReviewerRepository manages reviewers in the database
type ReviewerRepository struct {
	pool *pgxpool.Pool
}

// AssignReviewer assigns a reviewer to a PR
func (r *ReviewerRepository) AssignReviewer(ctx context.Context, prID, reviewerID string) error {
	query := `INSERT INTO pr_reviewer (pr_id, reviewer_id) 
	          VALUES ($1, $2)
	          ON CONFLICT (pr_id, reviewer_id) DO NOTHING`

	executor := getTx(ctx, r.pool)
	_, err := executor.Exec(ctx, query, prID, reviewerID)
	if err != nil {
		return fmt.Errorf("failed to assign reviewer: %w", err)
	}

	return nil
}

// GetReviewers gets all reviewers assigned to a PR
func (r *ReviewerRepository) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	query := `SELECT reviewer_id FROM pr_reviewer WHERE pr_id = $1 ORDER BY reviewer_id`

	executor := getTx(ctx, r.pool)
	rows, err := executor.Query(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer rows.Close()

	var reviewerIDs []string
	for rows.Next() {
		var reviewerID string
		if err = rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		reviewerIDs = append(reviewerIDs, reviewerID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return reviewerIDs, nil
}

// GetPRsByReviewer gets all PRs assigned to a reviewer
func (r *ReviewerRepository) GetPRsByReviewer(ctx context.Context, reviewerID string) ([]string, error) {
	query := `SELECT pr_id FROM pr_reviewer WHERE reviewer_id = $1 ORDER BY pr_id`

	executor := getTx(ctx, r.pool)
	rows, err := executor.Query(ctx, query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs by reviewer: %w", err)
	}
	defer rows.Close()

	var prIDs []string
	for rows.Next() {
		var prID string
		if err = rows.Scan(&prID); err != nil {
			return nil, fmt.Errorf("failed to scan PR ID: %w", err)
		}
		prIDs = append(prIDs, prID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return prIDs, nil
}

// IsAssigned checks if a reviewer is assigned to a PR
func (r *ReviewerRepository) IsAssigned(ctx context.Context, prID, reviewerID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pr_reviewer WHERE pr_id = $1 AND reviewer_id = $2)`

	executor := getTx(ctx, r.pool)
	var exists bool
	err := executor.QueryRow(ctx, query, prID, reviewerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check reviewer assignment: %w", err)
	}

	return exists, nil
}

// ReplaceReviewer replaces an old reviewer with a new one for a PR
func (r *ReviewerRepository) ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	executor := getTx(ctx, r.pool)

	deleteQuery := `DELETE FROM pr_reviewer WHERE pr_id = $1 AND reviewer_id = $2`
	_, err := executor.Exec(ctx, deleteQuery, prID, oldReviewerID)
	if err != nil {
		return fmt.Errorf("failed to remove old reviewer: %w", err)
	}

	insertQuery := `INSERT INTO pr_reviewer (pr_id, reviewer_id) VALUES ($1, $2)`
	_, err = executor.Exec(ctx, insertQuery, prID, newReviewerID)
	if err != nil {
		return fmt.Errorf("failed to assign new reviewer: %w", err)
	}

	return nil
}
