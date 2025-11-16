package pullrequest

// ReassignReviewerRequest represents a request to reassign a reviewer from a pull request.
type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
	OldReviewerID string `json:"old_reviewer_id" validate:"required"`
}

// ReassignReviewerResponse represents the response of reassigning a reviewer.
type ReassignReviewerResponse struct {
	Pr         PR     `json:"pr"`
	ReplacedBy string `json:"replaced_by"`
}
