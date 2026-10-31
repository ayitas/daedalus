package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ayitas/daedalus/internal/llm"
)

// ProgressStage represents whether a pass is starting or completed.
type ProgressStage string

const (
	StageStarting  ProgressStage = "starting"
	StageCompleted ProgressStage = "completed"
)

// ProgressCallback informs caller of the current pass and status.
type ProgressCallback func(pass, total int, name, description string, stage ProgressStage)

// Engine orchestrates the 3-pass LLM pipeline and scaffolding emission.
type Engine struct {
	llmClient *llm.Client
	renderer  *TemplateRenderer
	version   string
}

// NewEngine creates an initialized generation engine.
func NewEngine(client *llm.Client, version string) (*Engine, error) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize markdown templates: %w", err)
	}

	return &Engine{
		llmClient: client,
		renderer:  renderer,
		version:   version,
	}, nil
}

// Run executes the complete 3-pass generation pipeline with streaming support and writes files.
func (e *Engine) Run(ctx context.Context, input UserInput, targetBaseDir string, notify ProgressCallback, onChunk llm.StreamCallback) (*GenerationResult, error) {
	if notify == nil {
		notify = func(int, int, string, string, ProgressStage) {}
	}

	// -------------------------------------------------------------
	// Pass 1: The Architect
	// -------------------------------------------------------------
	notify(1, 3, "The Architect", "Synthesizing domain model, tech stack & user journeys...", StageStarting)
	techStackChoice := input.TechStackOption
	if input.CustomTechStack != "" {
		techStackChoice = fmt.Sprintf("Custom: %s", input.CustomTechStack)
	}

	p1Prompt := llm.FormatArchitectUserPrompt(
		input.ProjectName,
		input.AppConcept,
		input.AppFormat,
		input.TargetUsers,
		techStackChoice,
	)

	// Independent timeout per pass (up to 5m)
	p1Ctx, p1Cancel := context.WithTimeout(ctx, 5*time.Minute)
	p1Raw, err := e.llmClient.GenerateJSONStream(p1Ctx, llm.ArchitectSystemPrompt, p1Prompt, onChunk)
	p1Cancel()
	if err != nil {
		return nil, fmt.Errorf("pass 1 (Architect) failed: %w", err)
	}

	var p1Result Pass1ArchitectResult
	if err := json.Unmarshal(p1Raw, &p1Result); err != nil {
		return nil, fmt.Errorf("failed to decode Pass 1 Architect JSON: %w (raw: %s)", err, string(p1Raw))
	}
	if p1Result.AppName == "" {
		p1Result.AppName = input.ProjectName
	}
	notify(1, 3, "The Architect", fmt.Sprintf("Architecture synthesized for '%s' (%d entities, %d user flows)", p1Result.AppName, len(p1Result.EntitySchemas), len(p1Result.UserFlows)), StageCompleted)

	// -------------------------------------------------------------
	// Pass 2: The Tech Lead
	// -------------------------------------------------------------
	notify(2, 3, "The Tech Lead", "Formulating directory layout, verify commands & guardrails...", StageStarting)
	p2Prompt := llm.FormatTechLeadUserPrompt(p1Raw)

	p2Ctx, p2Cancel := context.WithTimeout(ctx, 5*time.Minute)
	p2Raw, err := e.llmClient.GenerateJSONStream(p2Ctx, llm.TechLeadSystemPrompt, p2Prompt, onChunk)
	p2Cancel()
	if err != nil {
		return nil, fmt.Errorf("pass 2 (Tech Lead) failed: %w", err)
	}

	var p2Result Pass2TechLeadResult
	if err := json.Unmarshal(p2Raw, &p2Result); err != nil {
		return nil, fmt.Errorf("failed to decode Pass 2 Tech Lead JSON: %w (raw: %s)", err, string(p2Raw))
	}
	notify(2, 3, "The Tech Lead", fmt.Sprintf("Guardrails formulated (Test: '%s', %d directories planned)", p2Result.VerifyCommands.Test, len(p2Result.DirectoryLayout)), StageCompleted)

	// -------------------------------------------------------------
	// Pass 3: The Planner
	// -------------------------------------------------------------
	notify(3, 3, "The Planner", "Decomposing features into atomic tasks & verification steps...", StageStarting)
	p3Prompt := llm.FormatPlannerUserPrompt(p1Raw, p2Raw)

	p3Ctx, p3Cancel := context.WithTimeout(ctx, 6*time.Minute)
	p3Raw, err := e.llmClient.GenerateJSONStream(p3Ctx, llm.PlannerSystemPrompt, p3Prompt, onChunk)
	p3Cancel()
	if err != nil {
		return nil, fmt.Errorf("pass 3 (Planner) failed: %w", err)
	}

	var p3Result Pass3PlannerResult
	if err := json.Unmarshal(p3Raw, &p3Result); err != nil {
		return nil, fmt.Errorf("failed to decode Pass 3 Planner JSON: %w (raw: %s)", err, string(p3Raw))
	}

	// Recalculate total tasks if not accurately reported
	totalTasks := 0
	for _, phase := range p3Result.Phases {
		totalTasks += len(phase.Tasks)
	}
	if totalTasks > 0 {
		p3Result.TotalTasks = totalTasks
	}
	notify(3, 3, "The Planner", fmt.Sprintf("Roadmap created (%d phases, %d atomic tasks)", len(p3Result.Phases), p3Result.TotalTasks), StageCompleted)

	// -------------------------------------------------------------
	// Go-Side Compilation & Scaffolding
	// -------------------------------------------------------------
	notify(4, 4, "Scaffolding & Templates", "Rendering deterministic markdown files and initializing git...", StageStarting)

	harnessData := &HarnessData{
		Architect:   p1Result,
		TechLead:    p2Result,
		Planner:     p3Result,
		UserInput:   input,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05 MST"),
		ToolVersion: e.version,
	}

	contextMD, err := e.renderer.RenderContext(harnessData)
	if err != nil {
		return nil, fmt.Errorf("failed to render CONTEXT.md: %w", err)
	}

	agentsMD, err := e.renderer.RenderAgents(harnessData)
	if err != nil {
		return nil, fmt.Errorf("failed to render AGENTS.md: %w", err)
	}

	progressMD, err := e.renderer.RenderProgress(harnessData)
	if err != nil {
		return nil, fmt.Errorf("failed to render PROGRESS.md: %w", err)
	}

	// Destination directory
	projectDir := filepath.Join(targetBaseDir, input.ProjectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", projectDir, err)
	}

	// Write files
	filesToCreate := map[string]string{
		"CONTEXT.md":  contextMD,
		"AGENTS.md":   agentsMD,
		"PROGRESS.md": progressMD,
	}

	var createdFiles []string
	for filename, content := range filesToCreate {
		targetPath := filepath.Join(projectDir, filename)
		if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", targetPath, err)
		}
		createdFiles = append(createdFiles, filename)
	}

	// Automatic git init
	gitCmd := exec.CommandContext(ctx, "git", "init")
	gitCmd.Dir = projectDir
	_ = gitCmd.Run() // Git init failure is non-fatal if git is missing

	notify(4, 4, "Scaffolding & Templates", fmt.Sprintf("Scaffolded into '%s' (CONTEXT.md, AGENTS.md, PROGRESS.md, git init)", projectDir), StageCompleted)

	// Construct the Magic Prompt
	magicPrompt := fmt.Sprintf(
		"I have set up the complete Daedalus Agent Harness for '%s'.\n"+
			"Please read CONTEXT.md for system architecture, AGENTS.md for operational rules and verification commands, and PROGRESS.md for the phased roadmap.\n"+
			"Begin with Phase 1, implement TASK-01, verify it with '%s', and update PROGRESS.md to mark it complete. Let's begin!",
		input.ProjectName,
		p2Result.VerifyCommands.Test,
	)

	return &GenerationResult{
		ProjectDir:   projectDir,
		FilesCreated: createdFiles,
		MagicPrompt:  magicPrompt,
		TotalPhases:  len(p3Result.Phases),
		TotalTasks:   p3Result.TotalTasks,
	}, nil
}
