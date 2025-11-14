package team

// GetTeamResponse represents the response when getting a team.
type GetTeamResponse struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}
