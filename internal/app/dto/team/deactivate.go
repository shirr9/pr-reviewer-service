package team

type DeactivateTeamRequest struct {
	TeamName string `json:"team_name" validate:"required"`
}

type DeactivateTeamResponse struct {
	DeactivatedUsers int      `json:"deactivated_users"`
	ReassignedPRs    int      `json:"reassigned_prs"`
	UserIDs          []string `json:"user_ids"`
}
