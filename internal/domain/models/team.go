package models

type User struct {
	Id       string
	Name     string
	TeamName string
	IsActive bool
}

// Team represent team members
type Team struct {
	Members []*User
}

// GetTeamName returns team name(from the first member)
// All the members have the same TeamName
func (t *Team) GetTeamName() string {
	if len(t.Members) == 0 {
		return ""
	}
	return t.Members[0].TeamName
}
