package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AIConfig holds settings for the AI suggestions feature.
type AIConfig struct {
	Provider     string `json:"provider"`       // "ollama" or "openai"
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key"`
	Model        string `json:"model"`
	SystemPrompt string `json:"system_prompt"`
}

// Config holds application configuration.
type Config struct {
	TabSize         int      `json:"tab_size"`
	WordWrap        bool     `json:"word_wrap"`
	ShowLineNumber  bool     `json:"show_line_numbers"`
	Theme           string   `json:"theme"`
	AutoSave        bool     `json:"auto_save"`
	AutoSaveDelay   int      `json:"auto_save_delay_ms"`
	BackgroundColor string   `json:"background_color"`
	GlamourStyle    string   `json:"glamour_style"`
	AI              AIConfig `json:"ai"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		TabSize:        4,
		WordWrap:       true,
		ShowLineNumber: true,
		Theme:          "auto",
		AutoSave:       false,
		AutoSaveDelay:  5000,
		GlamourStyle:   "dark",
		AI: AIConfig{
			Provider:     "ollama",
			BaseURL:      "http://localhost:11434",
			APIKey:       "",
			Model:        "llama3",
			SystemPrompt: "You are a helpful writing assistant. Suggest improvements or completions for the provided text. Be concise.",
		},
	}
}

// Load loads configuration from the default location.
func Load() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	return LoadFrom(configPath)
}

// LoadFrom loads configuration from a specific path.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// Save saves the configuration to the default location.
func (c *Config) Save() error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	return c.SaveTo(configPath)
}

// SaveTo saves the configuration to a specific path.
func (c *Config) SaveTo(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ConfigPath returns the default configuration file path.
func ConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "prose", "config.json"), nil
}
