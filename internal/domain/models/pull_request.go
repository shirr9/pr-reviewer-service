package models

import "time"

const (
	PRStatusOpen   = "OPEN"
	PRStatusMerged = "MERGED"
)

type PullRequest struct {
	Id          string
	Title       string
	AuthorId    string
	Status      string
	CreatedAt   time.Time
	MergedAt    *time.Time
	UpdatedAt   time.Time
	ReviewersId []string
}
