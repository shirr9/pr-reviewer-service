package pullrequest

// MergePrRequest represents a request to merge a pull request.
type MergePrRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
}

// MergePrResponse represents the response of merging a pull request.
type MergePrResponse struct {
	Pr PR `json:"pr"`
}
