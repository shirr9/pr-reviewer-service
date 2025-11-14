package user

// GetReviewResponse represents the response with user's assigned PRs for review.
type GetReviewResponse struct {
	UserID       string `json:"user_id"`
	PullRequests []PR   `json:"pull_requests"`
}

// PR represents short PR information.
type PR struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}
