package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/ayitas/daedalus/internal/config"
	"github.com/ayitas/daedalus/internal/generator"
	"github.com/ayitas/daedalus/internal/llm"
	"github.com/ayitas/daedalus/internal/ui"
)

const (
	AppVersion = "1.0.0"
	AppAuthor  = "ayitas"
)

func main() {
	var (
		flagVersion  = flag.Bool("version", false, "Print Daedalus version and exit")
		flagReconfig = flag.Bool("reconfig", false, "Re-run the provider & API credentials setup wizard")
		flagBaseDir  = flag.String("dir", ".", "Base directory where the project folder will be created")
	)
	flag.Parse()

	if *flagVersion {
		fmt.Printf("daedalus version %s by @%s (https://github.com/ayitas/daedalus)\n", AppVersion, AppAuthor)
		os.Exit(0)
	}

	// Print beautiful header banner with author credit
	fmt.Print(ui.RenderBanner(AppVersion, AppAuthor))

	// Step 1: Resolve Configuration
	var cfg *config.Config
	var err error

	if !*flagReconfig {
		cfg, err = config.LoadConfig()
	}

	if *flagReconfig || errors.Is(err, config.ErrNoConfigFound) || cfg == nil {
		if *flagReconfig {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.ColorSecondary).Render("⚙ Reconfiguring AI provider credentials..."))
		} else {
			fmt.Println(lipgloss.NewStyle().Foreground(ui.ColorWarning).Render("⚡ First-run detected: Configure your AI provider to begin."))
		}

		cfg, err = ui.RunSetupWizard()
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				fmt.Println("\nSetup cancelled.")
				os.Exit(0)
			}
			fmt.Fprintf(os.Stderr, "Error configuring provider: %v\n", err)
			os.Exit(1)
		}

		if saveErr := config.SaveConfig(cfg); saveErr != nil {
			fmt.Printf("Notice: could not save persistent config (%v), continuing with current session.\n", saveErr)
		} else {
			cfgPath, _ := config.GetConfigPath()
			fmt.Printf("Credentials saved securely to %s\n\n", cfgPath)
		}
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Status badge indicating active provider
	providerInfo := fmt.Sprintf("Active Provider: %s | Model: %s", strings.ToUpper(cfg.Provider), cfg.Model)
	fmt.Println(ui.BadgeStyle.Render(providerInfo))
	fmt.Println()

	// Step 2: Interactive Project Wizard
	userInput, err := ui.RunProjectWizard("my-app")
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Println("\nProject creation cancelled.")
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Wizard error: %v\n", err)
		os.Exit(1)
	}

	// Step 3: Run 3-Pass Generator Pipeline
	llmClient := llm.NewClient(cfg.Provider, cfg.BaseURL, cfg.APIKey, cfg.Model)

	engine, err := generator.NewEngine(llmClient, AppVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize generator engine: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.ColorSecondary).
		Render("🚀 Starting Autonomous Agent Harness Pipeline..."))
	fmt.Println()

	var result *generator.GenerationResult
	var generationErr error

	// Generous 20-minute global timeout to accommodate complex projects or local models
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	streamer := ui.NewThinkingStreamer()

	result, generationErr = engine.Run(
		ctx,
		*userInput,
		*flagBaseDir,
		func(pass, total int, name, description string, stage generator.ProgressStage) {
			if stage == generator.StageStarting {
				streamer.Reset()
				fmt.Printf("  [%d/%d] %s: %s\n", pass, total, lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary).Render(name), description)
			} else if stage == generator.StageCompleted {
				streamer.Close()
				fmt.Printf("  %s %s\n\n", lipgloss.NewStyle().Bold(true).Foreground(ui.ColorSuccess).Render("✔"), description)
			}
		},
		func(chunk string) {
			streamer.Feed(chunk)
		},
	)

	if generationErr != nil {
		streamer.Close()
		fmt.Fprintf(os.Stderr, "\nGeneration Error: %v\n", generationErr)
		os.Exit(1)
	}

	// Step 4: Display Completion Summary & Magic Prompt
	absDir, _ := filepath.Abs(result.ProjectDir)
	fmt.Println(ui.RenderSuccessCard(
		userInput.ProjectName,
		absDir,
		result.TotalPhases,
		result.TotalTasks,
		result.FilesCreated,
	))

	fmt.Println(ui.RenderMagicPrompt(result.MagicPrompt, absDir))
}
