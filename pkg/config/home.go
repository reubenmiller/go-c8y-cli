package config

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/reubenmiller/go-c8y-cli/pkg/fileutilities"
)

var (
	DefaultHome       = "~/.go-c8y-cli"
	DefaultSessionDir = "sessions"
	EnvHome           = "C8Y_HOME"
	EnvSessionHome    = "C8Y_SESSION_HOME"
)

func (c *Config) GetSessionHomeDir() string {
	var outputDir string

	if v := os.Getenv(EnvSessionHome); v != "" {
		expandedV, err := homedir.Expand(v)
		outputDir = v
		if err == nil {
			outputDir = expandedV
		} else if c.Logger != nil {
			c.Logger.Warnf("Could not expand path. %s", err)
		}
	} else {
		outputDir = c.GetHomeDir()
		outputDir = filepath.Join(outputDir, DefaultSessionDir)
	}

	err := fileutilities.CreateDirs(outputDir)
	if err != nil && c.Logger != nil {
		c.Logger.Errorf("Could not create sessions directory and it does not exist. %s", err)
	}
	return outputDir
}

// GetHomeDir get the home directory related to the cli tool
func (c *Config) GetHomeDir() string {
	outputDir := DefaultHome
	if v := os.Getenv(EnvHome); v != "" {
		outputDir = v
	}

	outputDir, err := homedir.Expand(outputDir)
	if err != nil && c.Logger != nil {
		c.Logger.Warnf("Could not expand path. %s", err)
	}

	err = fileutilities.CreateDirs(outputDir)
	if err != nil && c.Logger != nil {
		c.Logger.Errorf("Could not create sessions directory and it does not exist. %s", err)
	}
	return outputDir
}
