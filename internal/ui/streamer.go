package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ThinkingStreamer manages live terminal streaming of AI reasoning,
// displaying human-friendly natural language and preventing raw JSON syntax from flooding the screen.
type ThinkingStreamer struct {
	buffer           strings.Builder
	boxOpened        bool
	jsonMode         bool
	hasPrintedJSONMsg bool
	seenKeys         map[string]bool
}

// NewThinkingStreamer creates an initialized ThinkingStreamer.
func NewThinkingStreamer() *ThinkingStreamer {
	return &ThinkingStreamer{
		seenKeys: make(map[string]bool),
	}
}

// Reset resets state for the next pass.
func (s *ThinkingStreamer) Reset() {
	s.buffer.Reset()
	s.boxOpened = false
	s.jsonMode = false
	s.hasPrintedJSONMsg = false
	s.seenKeys = make(map[string]bool)
}

// Feed receives a raw token chunk from the LLM stream.
func (s *ThinkingStreamer) Feed(chunk string) {
	s.buffer.WriteString(chunk)
	full := s.buffer.String()

	// Detect transition to JSON block (either ```json, ``` or opening '{')
	if !s.jsonMode {
		// If chunk or buffer has ```json or starts with '{'
		if strings.Contains(chunk, "```json") || strings.Contains(chunk, "```") || strings.HasPrefix(strings.TrimSpace(full), "{") {
			s.jsonMode = true
			if s.boxOpened {
				fmt.Print(lipgloss.NewStyle().Foreground(ColorMuted).Render("\n  └───────────────────────────────────────────────────────────\n"))
				s.boxOpened = false
			}
			if !s.hasPrintedJSONMsg {
				fmt.Println(lipgloss.NewStyle().Foreground(ColorSecondary).Render("  ⚡ Compiling verified architecture contract into Agent Harness..."))
				s.hasPrintedJSONMsg = true
			}
			return
		}

		// Otherwise, this is natural conversational thinking or markdown explanation
		// Filter out any literal <thinking> or </thinking> tags if present
		cleanChunk := strings.ReplaceAll(chunk, "<thinking>", "")
		cleanChunk = strings.ReplaceAll(cleanChunk, "</thinking>", "")

		if strings.TrimSpace(cleanChunk) != "" || s.boxOpened {
			if !s.boxOpened {
				fmt.Print(lipgloss.NewStyle().Foreground(ColorMuted).Render("  ┌─ 💭 AI Architectural Reasoning ────────────────────────────\n  │ "))
				s.boxOpened = true
			}

			// Format newlines with left border indentation
			formatted := strings.ReplaceAll(cleanChunk, "\n", "\n  │ ")
			fmt.Print(lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4")).Render(formatted))
		}
	} else {
		// In JSON mode: provide live progress indicators for detected keys instead of dumping raw braces
		s.checkJSONKeys(chunk)
	}
}

// checkJSONKeys looks for key sections in the JSON stream to provide live, friendly status updates.
func (s *ThinkingStreamer) checkJSONKeys(chunk string) {
	keyMap := map[string]string{
		"problem_statement":      "Analyzing user problem & constraints...",
		"recommended_tech_stack": "Evaluating and selecting modern tech stack...",
		"entity_schemas":         "Designing domain entities & data schemas...",
		"user_flows":             "Mapping end-to-end user journeys...",
		"directory_layout":       "Structuring project file tree & directories...",
		"verify_commands":        "Configuring verification test & lint commands...",
		"coding_rules":           "Establishing strict non-negotiable coding rules...",
		"prohibitions":           "Defining strict agent prohibitions...",
		"phases":                 "Decomposing project into phased atomic tasks...",
	}

	for key, label := range keyMap {
		if !s.seenKeys[key] && strings.Contains(chunk, `"`+key+`"`) {
			s.seenKeys[key] = true
			fmt.Printf("    • %s\n", lipgloss.NewStyle().Foreground(ColorSecondary).Render(label))
		}
	}
}

// Close cleanly terminates any open UI boxes.
func (s *ThinkingStreamer) Close() {
	if s.boxOpened {
		fmt.Print(lipgloss.NewStyle().Foreground(ColorMuted).Render("\n  └───────────────────────────────────────────────────────────\n"))
		s.boxOpened = false
	}
}
