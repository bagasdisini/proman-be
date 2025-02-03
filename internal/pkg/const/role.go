package _const

// Role type
const (
	RoleAdmin      = "admin"
	RoleMaintainer = "maintainer"
	RoleDeveloper  = "developer"
)

// Position type
const (
	PositionCEO              = "Chief Executive Officer (CEO)"
	PositionCDBO             = "Chief Data and Business Officer (CDBO)"
	PositionCTO              = "Chief Technology Officer (CTO)"
	PositionCOO              = "Chief Operating Officer (COO)"
	PositionHRManager        = "HR/GA Manager"
	PositionLeadEngineer     = "Lead of Engineering"
	PositionLeadBA           = "Lead of Business and Analytics"
	PositionLeadContent      = "Lead of Content Analyst"
	PositionHROfficer        = "HR/GA Officer"
	PositionFE               = "Frontend Engineer"
	PositionBE               = "Backend Engineer"
	PositionQA               = "Quality Assurance Engineer"
	PositionDevOps           = "DevOps Engineer"
	PositionCrawler          = "Crawler Engineer"
	PositionTechSupport      = "Technical Support Engineer"
	PositionUIUX             = "UI/UX Engineer"
	PositionDataScientist    = "Data Scientist"
	PositionBusinessDev      = "Business Development"
	PositionCS               = "Customer Support"
	PositionSocMedSpecialist = "Social Media Specialist"
	PositionSocMedAdmin      = "Social Media Admin"
	PositionGraphicDesigner  = "Graphic Designer"
	PositionVideoEditor      = "Video Editor"
	PositionContentAnalyst   = "Content Analyst"
	PositionOther            = "Other"
)

func GetAllPositions() []string {
	return []string{
		PositionCEO,
		PositionCDBO,
		PositionCTO,
		PositionCOO,
		PositionHRManager,
		PositionLeadEngineer,
		PositionLeadBA,
		PositionLeadContent,
		PositionHROfficer,
		PositionFE,
		PositionBE,
		PositionQA,
		PositionDevOps,
		PositionCrawler,
		PositionTechSupport,
		PositionUIUX,
		PositionDataScientist,
		PositionBusinessDev,
		PositionCS,
		PositionSocMedSpecialist,
		PositionSocMedAdmin,
		PositionGraphicDesigner,
		PositionVideoEditor,
		PositionContentAnalyst,
		PositionOther,
	}
}

func IsValidPosition(position string) bool {
	for _, p := range GetAllPositions() {
		if p == position {
			return true
		}
	}
	return false
}
