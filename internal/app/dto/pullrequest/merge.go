package pullrequest

// MergePrRequest represents a request to merge a pull request.
type MergePrRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

// MergePrResponse represents the response of merging a pull request.
type MergePrResponse struct {
	Pr PR `json:"pr"`
}
