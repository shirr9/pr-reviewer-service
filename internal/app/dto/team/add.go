package team

// AddTeamRequest represents the request to create a team with members.
type AddTeamRequest struct {
	TeamName string       `json:"team_name" validate:"required"`
	Members  []TeamMember `json:"members" validate:"required,min=1,dive"`
}

// TeamMember represents a member of the team.
type TeamMember struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required"`
	IsActive bool   `json:"is_active"`
}

// AddTeamResponse represents the response after creating a team.
type AddTeamResponse struct {
	Team Team `json:"team"`
}

// Team represents team data with members.
type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}
