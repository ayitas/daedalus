package ui

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"

	"github.com/ayitas/daedalus/internal/config"
	"github.com/ayitas/daedalus/internal/generator"
)

var projectNameRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// RunSetupWizard prompts the user to select an LLM provider and input necessary credentials.
func RunSetupWizard() (*config.Config, error) {
	fmt.Println(lipgloss.NewStyle().Foreground(ColorMuted).Render(
		"💡 Keyboard Controls: [Tab / ↓] Next • [Shift+Tab / ↑] Back • [Enter] Select • [Ctrl+C] Cancel",
	))
	fmt.Println()

	var selectedProvider string
	var apiKey string
	var customBaseURL string
	var customModel string

	providerForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose an AI Provider").
				Description("Select where Daedalus will route architecture synthesis calls").
				Options(
					huh.NewOption("Local Model (Ollama / LM Studio / LocalAI — Open-Weights, Free & Offline)", config.ProviderLocal),
					huh.NewOption("Google Gemini (Google AI Studio — Generous Free Tier API Key)", config.ProviderGemini),
					huh.NewOption("Anthropic Claude (Claude 3.5 / 3.7 Sonnet via OpenRouter API Key)", config.ProviderClaude),
					huh.NewOption("OpenAI Codex / GPT-4o (Official OpenAI API Key)", config.ProviderOpenAI),
					huh.NewOption("Custom Endpoint (Any OpenAI-compatible server or proxy)", config.ProviderCustom),
				).
				Value(&selectedProvider),
		),
	).WithShowHelp(true)

	if err := providerForm.Run(); err != nil {
		return nil, err
	}

	var authGroup *huh.Group

	switch selectedProvider {
	case config.ProviderLocal, config.ProviderOllama:
		customModel = config.DefaultLocalModel
		customBaseURL = config.DefaultLocalURL
		authGroup = huh.NewGroup(
			huh.NewInput().
				Title("Local Server Endpoint").
				Description("Default is http://localhost:11434/v1 (Ollama) or http://localhost:1234/v1 (LM Studio)").
				Value(&customBaseURL),
			huh.NewInput().
				Title("Model Name").
				Description("e.g. qwen2.5-coder:7b, deepseek-r1:8b, llama3.3:70b, mistral").
				Value(&customModel),
		)

	case config.ProviderGemini:
		customModel = config.DefaultGeminiModel
		customBaseURL = config.DefaultGeminiURL
		authGroup = huh.NewGroup(
			huh.NewInput().
				Title("Google AI Studio API Key").
				Description("Get your free Gemini API key from https://aistudio.google.com/app/apikey").
				EchoMode(huh.EchoModePassword).
				Value(&apiKey).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("API key cannot be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("Model Name (Optional)").
				Description(fmt.Sprintf("Default is '%s' (e.g. gemini-3.8-flash, gemini-3.8-pro)", config.DefaultGeminiModel)).
				Value(&customModel),
		)

	case config.ProviderClaude, config.ProviderOpenRouter:
		customModel = config.DefaultClaudeModel
		customBaseURL = config.DefaultClaudeURL
		authGroup = huh.NewGroup(
			huh.NewInput().
				Title("Claude API Key (via OpenRouter)").
				Description("Get your key from https://openrouter.ai/keys to use Claude without region blocks").
				EchoMode(huh.EchoModePassword).
				Value(&apiKey).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("API key cannot be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("Model (Optional)").
				Description(fmt.Sprintf("Default is '%s' (e.g. anthropic/claude-3.7-sonnet)", config.DefaultClaudeModel)).
				Value(&customModel),
		)

	case config.ProviderOpenAI:
		customModel = config.DefaultOpenAIModel
		customBaseURL = config.DefaultOpenAIURL
		authGroup = huh.NewGroup(
			huh.NewInput().
				Title("OpenAI API Key").
				Description("Get key from https://platform.openai.com/api-keys for GPT-4o / Codex").
				EchoMode(huh.EchoModePassword).
				Value(&apiKey).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("API key cannot be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("Model Name (Optional)").
				Description(fmt.Sprintf("Default is '%s' (e.g. gpt-4o, o3-mini, o1)", config.DefaultOpenAIModel)).
				Value(&customModel),
		)

	default: // Custom
		authGroup = huh.NewGroup(
			huh.NewInput().
				Title("Custom Base URL").
				Description("OpenAI-compatible base URL (e.g., https://my-llm.corp.internal/v1)").
				Value(&customBaseURL).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("Base URL cannot be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("API Key").
				Description("Provider API Key or token (leave blank if not needed)").
				EchoMode(huh.EchoModePassword).
				Value(&apiKey),
			huh.NewInput().
				Title("Model Name").
				Description("Target model name identifier").
				Value(&customModel).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("Model name cannot be empty")
					}
					return nil
				}),
		)
	}

	authForm := huh.NewForm(authGroup).WithShowHelp(true)
	if err := authForm.Run(); err != nil {
		return nil, err
	}

	cfg := config.PresetForProvider(selectedProvider, apiKey, customModel)
	if customBaseURL != "" {
		cfg.BaseURL = customBaseURL
	}
	if customModel != "" {
		cfg.Model = customModel
	}

	return cfg, nil
}

// RunProjectWizard prompts the non-technical founder for plain-language product specifications.
func RunProjectWizard(defaultName string) (*generator.UserInput, error) {
	fmt.Println(lipgloss.NewStyle().Foreground(ColorMuted).Render(
		"💡 Keyboard Controls: [Tab / ↓] Next Field • [Shift+Tab / ↑] Previous Field (Back) • [Enter] Confirm • [Ctrl+C] Quit",
	))
	fmt.Println()

	var (
		projectName     = defaultName
		appConcept      string
		appFormat       string
		targetUsers     string
		techStackOption string
		customTechStack string
	)

	projectGroup := huh.NewGroup(
		huh.NewInput().
			Title("Project Name").
			Description("Folder/repository name (letters, numbers, and hyphens only)").
			Value(&projectName).
			Validate(func(s string) error {
				trimmed := strings.TrimSpace(s)
				if trimmed == "" {
					return errors.New("project name cannot be empty")
				}
				if !projectNameRegex.MatchString(trimmed) {
					return errors.New("project name may only contain letters, numbers, and hyphens (no spaces)")
				}
				return nil
			}),

		huh.NewText().
			Title("App Concept").
			Description("What problem does this solve, and what should happen when someone uses it?").
			Lines(4).
			Value(&appConcept).
			Validate(func(s string) error {
				if len(strings.TrimSpace(s)) < 10 {
					return errors.New("please describe your concept in at least 10 characters")
				}
				return nil
			}),

		huh.NewSelect[string]().
			Title("App Format").
			Description("What kind of application do you want to build?").
			Options(
				huh.NewOption("Web App (Browser based)", "Web App"),
				huh.NewOption("Messaging Bot (Telegram / Discord)", "Messaging Bot (Telegram/Discord)"),
				huh.NewOption("CLI Tool (Terminal based)", "CLI Tool"),
				huh.NewOption("Mobile-friendly Web (Responsive PWA)", "Mobile-friendly Web"),
				huh.NewOption("Let AI decide the best format", "Let AI decide"),
			).
			Value(&appFormat),

		huh.NewSelect[string]().
			Title("Target Users").
			Description("Who will be using this application?").
			Options(
				huh.NewOption("Just me (Single user / Local / Offline)", "Just me (Single user/Offline)"),
				huh.NewOption("Public users (Multi-user with Authentication)", "Public users (Multi-user/Auth)"),
				huh.NewOption("Team tool (Internal company / Workspace)", "Team tool"),
			).
			Value(&targetUsers),

		huh.NewSelect[string]().
			Title("Technology Stack").
			Description("Choose how technologies should be selected").
			Options(
				huh.NewOption("[Recommended] Simplest & modern (Auto-selected by AI Architect)", "[Recommended] Simplest & modern (Auto)"),
				huh.NewOption("I want to specify manually", "I want to specify manually"),
			).
			Value(&techStackOption),
	)

	form := huh.NewForm(projectGroup).WithShowHelp(true)
	if err := form.Run(); err != nil {
		return nil, err
	}

	// If manual tech stack was chosen, ask for preferences
	if techStackOption == "I want to specify manually" {
		manualForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Specify Preferred Tech Stack").
					Description("e.g. Next.js + PostgreSQL + Prisma or Go + HTMX + SQLite").
					Value(&customTechStack).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return errors.New("please enter your preferred tech stack")
						}
						return nil
					}),
			),
		).WithShowHelp(true)

		if err := manualForm.Run(); err != nil {
			return nil, err
		}
	}

	return &generator.UserInput{
		ProjectName:     strings.TrimSpace(projectName),
		AppConcept:      strings.TrimSpace(appConcept),
		AppFormat:       appFormat,
		TargetUsers:     targetUsers,
		TechStackOption: techStackOption,
		CustomTechStack: strings.TrimSpace(customTechStack),
	}, nil
}

// RunSpinner runs an action function while displaying a stylized spinner.
func RunSpinner(title string, action func() error) error {
	var runErr error
	err := spinner.New().
		Title(title).
		Action(func() {
			runErr = action()
		}).
		Run()

	if err != nil {
		return err
	}
	return runErr
}
