package config

import (
	"fmt"
	"os"
	"path/filepath"
	"soloterm/domain/tag"
	"soloterm/shared/validation"

	"gopkg.in/yaml.v3"
)

const (
	CONFIG_FILE_NAME = "config.yaml"
)

// Config represents the application configuration
type Config struct {
	FullFilePath    string        `yaml:"-"`
	TagTypes        []tag.TagType `yaml:"tag_types"`
	TagExcludeWords []string      `yaml:"tag_exclude_words"`
}

// Load loads the configuration file from the directory passed in
func (c *Config) Load(workdir string) (*Config, error) {
	c.FullFilePath = filepath.Join(workdir, CONFIG_FILE_NAME)
	var data []byte

	// If the file doesn't exist, create it, otherwise read it
	if c.fileExists(c.FullFilePath) == false {
		if err := c.writeDefault(c.FullFilePath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}

	// Read the file
	data, err := os.ReadFile(c.FullFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.FullFilePath = c.FullFilePath

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.TagTypes) == 0 {
		return fmt.Errorf("tag_types cannot be empty")
	}

	var v *validation.Validator

	for i := range c.TagTypes {
		v = c.TagTypes[i].Validate()
		if v.HasError("label") {
			return fmt.Errorf("tag_types[%d]: label is required", i)
		}
		if v.HasError("template") {
			return fmt.Errorf("tag_types[%d]: template is required", i)
		}
	}

	return nil
}

// writeDefault writes the default configuration file
func (c *Config) writeDefault(filepath string) error {
	cfg := Config{
		TagTypes:        tag.DefaultTagTypes(),
		TagExcludeWords: []string{"closed", "abandoned"},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add comments to the YAML
	yamlWithComments := `# Soloterm Configuration
#
# Tag Types define the Lonelog notation tags available in the app.
# Each tag type has:
#   label:    The human-readable name shown in the UI
#   template: The Lonelog notation pattern inserted when selected
#
# Standard Lonelog tag types are provided below.
# Add, remove, or modify entries to suit your game system.
#
# Tag Exclude Words are terms that, when found in the data section of a tag,
# will exclude that tag from appearing in the recent tags list.
# This is useful for filtering out completed or archived tags.
# Words are matched case-insensitively.

` + string(data)

	// Write to file
	if err := os.WriteFile(filepath, []byte(yamlWithComments), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) fileExists(filepath string) bool {
	info, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	}
	return info != nil && !info.IsDir()
}
