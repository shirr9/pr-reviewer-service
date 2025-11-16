package statistics

type UserStats struct {
	UserID           string `json:"user_id"`
	Username         string `json:"username"`
	AssignmentsCount int    `json:"assignments_count"`
	ActiveReviews    int    `json:"active_reviews"`
}

type PRStats struct {
	PullRequestID      string `json:"pull_request_id"`
	PullRequestName    string `json:"pull_request_name"`
	ReviewersCount     int    `json:"reviewers_count"`
	Status             string `json:"status"`
	ReassignmentsCount int    `json:"reassignments_count"`
}

type StatisticsResponse struct {
	TotalPRs         int         `json:"total_prs"`
	OpenPRs          int         `json:"open_prs"`
	MergedPRs        int         `json:"merged_prs"`
	TotalAssignments int         `json:"total_assignments"`
	UserStats        []UserStats `json:"user_stats,omitempty"`
	PRStats          []PRStats   `json:"pr_stats,omitempty"`
}
