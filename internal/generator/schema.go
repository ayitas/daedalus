package generator

// UserInput captures the non-technical founder's inputs collected by the TUI wizard.
type UserInput struct {
	ProjectName     string `json:"project_name"`
	AppConcept      string `json:"app_concept"`
	AppFormat       string `json:"app_format"`
	TargetUsers     string `json:"target_users"`
	TechStackOption string `json:"tech_stack_option"`
	CustomTechStack string `json:"custom_tech_stack,omitempty"`
}

// TechStackDef specifies the architectural components recommended by the Architect.
type TechStackDef struct {
	Frontend  string `json:"frontend"`
	Backend   string `json:"backend"`
	Database  string `json:"database"`
	Auth      string `json:"auth"`
	Styling   string `json:"styling"`
	Rationale string `json:"rationale"`
}

// FieldDef defines a single field within an entity.
type FieldDef struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// EntityDef describes a domain data model/schema.
type EntityDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Fields      []FieldDef `json:"fields"`
}

// UserFlowDef represents an end-to-end user journey.
type UserFlowDef struct {
	FlowName    string   `json:"flow_name"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
}

// Pass1ArchitectResult is the contract output of the Architect pass.
type Pass1ArchitectResult struct {
	AppName              string        `json:"app_name"`
	ProblemStatement     string        `json:"problem_statement"`
	TargetAudience       string        `json:"target_audience"`
	RecommendedTechStack TechStackDef  `json:"recommended_tech_stack"`
	EntitySchemas        []EntityDef   `json:"entity_schemas"`
	UserFlows            []UserFlowDef `json:"user_flows"`
}

// DirItem describes a file or directory in the proposed layout.
type DirItem struct {
	Path    string `json:"path"`
	Purpose string `json:"purpose"`
}

// VerifyCommandsDef captures executable commands to validate the codebase.
type VerifyCommandsDef struct {
	Setup string `json:"setup"`
	Test  string `json:"test"`
	Lint  string `json:"lint"`
	Run   string `json:"run"`
	Build string `json:"build"`
}

// Pass2TechLeadResult is the contract output of the Tech Lead pass.
type Pass2TechLeadResult struct {
	DirectoryLayout  []DirItem         `json:"directory_layout"`
	VerifyCommands   VerifyCommandsDef `json:"verify_commands"`
	CodingRules      []string          `json:"coding_rules"`
	Prohibitions     []string          `json:"prohibitions"`
	AgentConventions []string          `json:"agent_conventions"`
}

// TaskDef defines an atomic implementation task for an autonomous coding agent.
type TaskDef struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	VerificationStep string `json:"verification_step"`
	Status           string `json:"status"` // always "pending" initially
}

// PhaseDef groups atomic tasks into chronological implementation phases.
type PhaseDef struct {
	PhaseNumber int       `json:"phase_number"`
	PhaseName   string    `json:"phase_name"`
	Goal        string    `json:"goal"`
	Tasks       []TaskDef `json:"tasks"`
}

// Pass3PlannerResult is the contract output of the Planner pass.
type Pass3PlannerResult struct {
	Phases     []PhaseDef `json:"phases"`
	TotalTasks int        `json:"total_tasks"`
}

// HarnessData is the combined compilation context used by Go's text/template engine.
type HarnessData struct {
	Architect   Pass1ArchitectResult
	TechLead    Pass2TechLeadResult
	Planner     Pass3PlannerResult
	UserInput   UserInput
	GeneratedAt string
	ToolVersion string
}

// GenerationResult represents the output of the full 3-pass generation pipeline.
type GenerationResult struct {
	ProjectDir   string
	FilesCreated []string
	MagicPrompt  string
	TotalPhases  int
	TotalTasks   int
}
