package service

// Role defines the user role
type Role string

// GetName gets the name of a role
func (r Role) GetName() string {
	return string(r)
}

func ValidateRole(r Role) bool {
	for _, ar := range AllowedRoles {
		if ar == r {
			return true
		}
	}
	return false
}

// ListAllowedRoleName lists names of all allowed roles
func ListAllowedRoleName() []string {
	var roleNames []string
	for _, ar := range AllowedRoles {
		roleNames = append(roleNames, ar.GetName())
	}
	return roleNames
}

// Define roles
const (
	Manager           Role = "manager"
	AIResearcher      Role = "ai-researcher"
	AIEngineer        Role = "ai-engineer"
	DataEngineer      Role = "data-engineer"
	DataScientist     Role = "data-scientist"
	AnalyticsEngineer Role = "analytics-engineer"
	Hobbyist          Role = "hobbyist"
)

var AllowedRoles = []Role{
	Manager,
	AIResearcher,
	AIEngineer,
	DataEngineer,
	DataScientist,
	AnalyticsEngineer,
	Hobbyist,
}
