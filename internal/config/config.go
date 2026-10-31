package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Supported providers
const (
	ProviderLocal      = "local"
	ProviderOllama     = "ollama" // alias for local
	ProviderGemini     = "gemini"
	ProviderClaude     = "claude"
	ProviderOpenRouter = "openrouter" // alias for claude / openrouter
	ProviderOpenAI     = "openai"
	ProviderCustom     = "custom"
)

// Default endpoints and models
const (
	DefaultLocalURL        = "http://localhost:11434/v1"
	DefaultLocalModel      = "qwen2.5-coder:7b"
	DefaultOllamaURL       = "http://localhost:11434/v1"
	DefaultOllamaTagsURL   = "http://localhost:11434/api/tags"
	DefaultOllamaModel     = "qwen2.5-coder:7b"
	DefaultGeminiURL       = "https://generativelanguage.googleapis.com/v1beta/openai"
	DefaultGeminiModel     = "gemini-3.8-flash"
	DefaultClaudeURL       = "https://openrouter.ai/api/v1"
	DefaultClaudeModel     = "anthropic/claude-3.5-sonnet"
	DefaultOpenRouterURL   = DefaultClaudeURL
	DefaultOpenRouterModel = DefaultClaudeModel
	DefaultOpenAIURL       = "https://api.openai.com/v1"
	DefaultOpenAIModel     = "gpt-4o"
	ConfigDirName          = "daedalus"
	ConfigFileName         = "config.json"
)

var (
	ErrNoConfigFound = errors.New("no saved configuration or active credentials found")
)

// Config holds the LLM connectivity and provider settings.
type Config struct {
	Provider string `json:"provider"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
}

// OllamaTagsResponse represents the response from /api/tags
type OllamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// GetConfigPath returns the absolute path to ~/.config/daedalus/config.json.
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, hErr := os.UserHomeDir()
		if hErr != nil {
			return "", fmt.Errorf("failed to determine user home or config directory: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, ConfigDirName, ConfigFileName), nil
}

// DetectLocalOllama attempts to ping Ollama's local tags API.
// If reachable, it returns true and the best detected model name (or default).
func DetectLocalOllama(timeout time.Duration) (bool, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, DefaultOllamaTagsURL, nil)
	if err != nil {
		return false, "", err
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var tagsResp OllamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return true, DefaultOllamaModel, nil
	}

	if len(tagsResp.Models) == 0 {
		return true, DefaultOllamaModel, nil
	}

	// Prefer coder models if available
	for _, m := range tagsResp.Models {
		lower := strings.ToLower(m.Name)
		if strings.Contains(lower, "coder") || strings.Contains(lower, "qwen") {
			return true, m.Name, nil
		}
	}

	// Otherwise return the first available model
	return true, tagsResp.Models[0].Name, nil
}

// LoadConfig tries to locate configuration in the following order:
// 1. Environment variables (DAEDALUS_API_KEY, DAEDALUS_BASE_URL, DAEDALUS_MODEL, DAEDALUS_PROVIDER)
// 2. Persistent config file (~/.config/daedalus/config.json)
// 3. Local Ollama instance detection (auto zero-friction setup)
func LoadConfig() (*Config, error) {
	// 1. Check environment variables
	if envKey := strings.TrimSpace(os.Getenv("DAEDALUS_API_KEY")); envKey != "" {
		provider := os.Getenv("DAEDALUS_PROVIDER")
		baseURL := os.Getenv("DAEDALUS_BASE_URL")
		model := os.Getenv("DAEDALUS_MODEL")

		if provider == "" {
			provider = ProviderCustom
		}
		if baseURL == "" {
			baseURL = DefaultOpenAIURL
		}
		if model == "" {
			model = DefaultOpenAIModel
		}

		return &Config{
			Provider: provider,
			BaseURL:  baseURL,
			APIKey:   envKey,
			Model:    model,
		}, nil
	}

	// Check generic provider env vars
	if key := strings.TrimSpace(os.Getenv("GEMINI_API_KEY")); key != "" {
		return &Config{
			Provider: ProviderGemini,
			BaseURL:  DefaultGeminiURL,
			APIKey:   key,
			Model:    DefaultGeminiModel,
		}, nil
	}

	if key := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY")); key != "" {
		return &Config{
			Provider: ProviderClaude,
			BaseURL:  DefaultClaudeURL,
			APIKey:   key,
			Model:    DefaultClaudeModel,
		}, nil
	}

	if key := strings.TrimSpace(os.Getenv("CLAUDE_API_KEY")); key != "" {
		return &Config{
			Provider: ProviderClaude,
			BaseURL:  DefaultClaudeURL,
			APIKey:   key,
			Model:    DefaultClaudeModel,
		}, nil
	}

	if key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); key != "" {
		return &Config{
			Provider: ProviderOpenAI,
			BaseURL:  DefaultOpenAIURL,
			APIKey:   key,
			Model:    DefaultOpenAIModel,
		}, nil
	}

	// 2. Check persistent config file
	cfgPath, err := GetConfigPath()
	if err == nil {
		if data, readErr := os.ReadFile(cfgPath); readErr == nil {
			var cfg Config
			if jsonErr := json.Unmarshal(data, &cfg); jsonErr == nil && cfg.BaseURL != "" {
				if cfg.Provider == ProviderOllama && cfg.APIKey == "" {
					cfg.APIKey = "ollama"
				}
				// Auto-migrate deprecated Gemini models to the latest supported model
				if cfg.Provider == ProviderGemini && (cfg.Model == "gemini-2.5-flash" || cfg.Model == "") {
					cfg.Model = DefaultGeminiModel
					_ = SaveConfig(&cfg)
				}
				if cfg.APIKey != "" || cfg.Provider == ProviderOllama || cfg.Provider == ProviderLocal {
					return &cfg, nil
				}
			}
		}
	}

	// 3. Auto-detect local Ollama
	if running, detectedModel, err := DetectLocalOllama(1200 * time.Millisecond); err == nil && running {
		return &Config{
			Provider: ProviderOllama,
			BaseURL:  DefaultOllamaURL,
			APIKey:   "ollama", // Dummy key required by OpenAI client format
			Model:    detectedModel,
		}, nil
	}

	return nil, ErrNoConfigFound
}

// SaveConfig saves the configuration to the user's config directory with 0600 permissions.
func SaveConfig(cfg *Config) error {
	cfgPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", cfgPath, err)
	}

	return nil
}

// PresetForProvider returns a preconfigured Config struct for a given provider name.
func PresetForProvider(provider string, apiKey string, customModel string) *Config {
	switch strings.ToLower(provider) {
	case ProviderLocal, ProviderOllama:
		model := DefaultLocalModel
		if customModel != "" {
			model = customModel
		}
		key := apiKey
		if key == "" {
			key = "local"
		}
		return &Config{
			Provider: ProviderLocal,
			BaseURL:  DefaultLocalURL,
			APIKey:   key,
			Model:    model,
		}

	case ProviderGemini:
		model := DefaultGeminiModel
		if customModel != "" {
			model = customModel
		}
		return &Config{
			Provider: ProviderGemini,
			BaseURL:  DefaultGeminiURL,
			APIKey:   apiKey,
			Model:    model,
		}

	case ProviderClaude, ProviderOpenRouter:
		model := DefaultClaudeModel
		if customModel != "" {
			model = customModel
		}
		return &Config{
			Provider: ProviderClaude,
			BaseURL:  DefaultClaudeURL,
			APIKey:   apiKey,
			Model:    model,
		}

	case ProviderOpenAI:
		model := DefaultOpenAIModel
		if customModel != "" {
			model = customModel
		}
		return &Config{
			Provider: ProviderOpenAI,
			BaseURL:  DefaultOpenAIURL,
			APIKey:   apiKey,
			Model:    model,
		}

	default:
		return &Config{
			Provider: ProviderCustom,
			BaseURL:  DefaultOpenAIURL,
			APIKey:   apiKey,
			Model:    DefaultOpenAIModel,
		}
	}
}
