package config

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/reubenmiller/go-c8y-cli/pkg/fileutilities"
)

var (
	DefaultHome = []string{
		"~/.go-c8y-cli",                   // user's home folder
		"/usr/local/etc/go-c8y-cli",       // Default homebrew prefix
		"$HOMEBREW_PREFIX/etc/go-c8y-cli", // Check custom homebrew prefix
		"/etc/go-c8y-cli",                 // default when installing via a package (not git)
	}
	DefaultSessionDir = ".cumulocity"
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
		// Use session directory ~/.cumulocity (separate from c8y home, as it can store sensitive information)
		// Otherwise if the home directory can't be found, use the current directory
		if v, err := homedir.Dir(); err == nil {
			outputDir = v
		} else {
			if c.Logger != nil {
				c.Logger.Warnf("Could not find user's home directory. %s", err)
			}
		}
		outputDir = filepath.Join(outputDir, DefaultSessionDir)
	}

	err := fileutilities.CreateDirs(outputDir)
	if err != nil && c.Logger != nil {
		c.Logger.Errorf("Sessions directory check failed. path=%s, err=%s", outputDir, err)
	}
	return outputDir
}

// GetHomeDir get the home directory related to the cli tool
func (c *Config) GetHomeDir() string {
	outputDir := ""
	if v := os.Getenv(EnvHome); v != "" {
		outputDir = v
	}

	// use first existing default home path
	if outputDir == "" {
		for _, p := range DefaultHome {
			p, _ = homedir.Expand(os.ExpandEnv(p))
			if stat, err := os.Stat(p); err == nil && stat.IsDir() {
				outputDir = p
				break
			}
		}
	}

	outputDir, err := homedir.Expand(os.ExpandEnv(outputDir))
	if err != nil && c.Logger != nil {
		c.Logger.Warnf("Could not expand path. %s", err)
	}

	err = fileutilities.CreateDirs(outputDir)
	if err != nil && c.Logger != nil {
		c.Logger.Errorf("Sessions directory check failed. path=%s, err=%s", outputDir, err)
	}
	return outputDir
}
