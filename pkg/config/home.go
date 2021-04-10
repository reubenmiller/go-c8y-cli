package config

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/reubenmiller/go-c8y-cli/pkg/fileutilities"
)

func (c *Config) GetSessionHomeDir() string {
	var outputDir string

	if v := os.Getenv("C8Y_SESSION_HOME"); v != "" {
		expandedV, err := homedir.Expand(v)
		outputDir = v
		if err == nil {
			outputDir = expandedV
		} else if c.Logger != nil {
			c.Logger.Warnf("Could not expand path. %s", err)
		}
	} else {
		outputDir, _ = homedir.Dir()
		outputDir = filepath.Join(outputDir, ".cumulocity", "sessions")
	}

	err := fileutilities.CreateDirs(outputDir)
	if err != nil && c.Logger != nil {
		c.Logger.Errorf("Could not create sessions directory and it does not exist. %s", err)
	}
	return outputDir
}
