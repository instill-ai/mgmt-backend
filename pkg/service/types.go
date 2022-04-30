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
	Manager           Role = "Manager"
	AIResearcher      Role = "AI Researcher"
	AIEngineer        Role = "AI Engineer"
	DataEngineer      Role = "Data Engineer"
	DataScientist     Role = "Data Scientist"
	AnalyticsEngineer Role = "Analytics Engineer"
	Hobbyist          Role = "Hobbyist"
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
