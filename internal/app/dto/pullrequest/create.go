package pullrequest

// PR represents info about a pull request.
type PR struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	MergedAt          string   `json:"mergedAt,omitempty"`
}

// CreatePrRequest represents a request to create a new pull request.
type CreatePrRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

// CreatePrResponse represents the response of creating a pull request.
type CreatePrResponse struct {
	Pr PR `json:"pr"`
}
