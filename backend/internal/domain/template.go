package domain

type Mode string

const (
	ModeGeneral Mode = "general"
	ModeSE      Mode = "se"
)

type Template struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Mode          Mode           `json:"mode"`
	WorkItemTypes []WorkItemType `json:"work_item_types"`
	Roles         []ProjectRole  `json:"roles"`
	DefaultRole   ProjectRole    `json:"default_role"`
	CreatorRole   ProjectRole    `json:"creator_role"`
	Capabilities  []string       `json:"capabilities"`
	// ValidationRules drive ValidateWorkItemAttributes for projects using this Template.
	ValidationRules ValidationRules `json:"validation_rules"`
}

const (
	TemplateIDGeneral = "general"
	TemplateIDSE      = "se"
)

func catalog() []Template {
	return []Template{
		{
			ID:            TemplateIDGeneral,
			Name:          "General",
			Mode:          ModeGeneral,
			WorkItemTypes: []WorkItemType{WorkItemTypeTask},
			Roles:         []ProjectRole{"owner", "member"},
			DefaultRole:   "member",
			CreatorRole:   "owner",
			Capabilities:  []string{},
			ValidationRules: ValidationRules{
				RequiredFields: []string{WorkItemFieldTitle},
				TypeRules:      []WorkItemTypeRule{},
			},
		},
		{
			ID:            TemplateIDSE,
			Name:          "Software Engineering",
			Mode:          ModeSE,
			WorkItemTypes: []WorkItemType{WorkItemTypeUserStory, WorkItemTypeTask, WorkItemTypeBug},
			Roles:         []ProjectRole{"product_owner", "scrum_master", "developer"},
			DefaultRole:   "developer",
			CreatorRole:   "product_owner",
			Capabilities:  []string{"user_story_guidance", "acceptance_criteria", "story_points", "sprints", "se_templates"},
			ValidationRules: ValidationRules{
				RequiredFields: []string{WorkItemFieldTitle},
				StoryPoints:    &StoryPointRule{GreaterThan: 0},
				TypeRules: []WorkItemTypeRule{{
					Type:                      WorkItemTypeUserStory,
					DescriptionFormat:         DescriptionFormatUserStory,
					RequireAcceptanceCriteria: true,
				}},
			},
		},
	}
}

// AvailableTemplates returns a fresh copy of the catalog on every call.
func AvailableTemplates() []Template {
	return catalog()
}

// TemplateByID returns a fresh copy of the template with the given id.
func TemplateByID(id string) (Template, bool) {
	for _, t := range catalog() {
		if t.ID == id {
			return t, true
		}
	}
	return Template{}, false
}
