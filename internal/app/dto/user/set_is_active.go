package user

// SetIsActiveRequest represents the request to set user's active status.
type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

// SetIsActiveResponse represents the response after setting user's active status.
type SetIsActiveResponse struct {
	User User `json:"user"`
}

// User represents user data
type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}
