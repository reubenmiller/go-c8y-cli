package models

import "strings"

type TestSuite struct {
	Configuration *Configuration      `yaml:"config,omitempty"`
	Tests         map[string]TestCase `yaml:"tests"`
}

type Configuration struct {
	Dir string            `yaml:"dir,omitempty"`
	Env map[string]string `yaml:"dir,env"`
}

type TestCase struct {
	Command       string           `yaml:"command,omitempty"`
	ExitCode      int              `yaml:"exit-code"`
	Skip          bool             `yaml:"skip,omitempty"`
	LineCount     int              `yaml:"line-count,omitempty"`
	StdOut        *OutputAssertion `yaml:"stdout,omitempty"`
	StdErr        *OutputAssertion `yaml:"stderr,omitempty"`
	Configuration *Configuration   `yaml:"config,omitempty"`
}

type OutputAssertion struct {
	JSON     map[string]string `yaml:"json,omitempty"`
	Contains []string          `yaml:"contains,omitempty"`
}

type MockConfiguration struct {
	Mocks map[string]string `yaml:"mocks,omitempty"`
	Files map[string]string `yaml:"files,omitempty"`
}

func (c *MockConfiguration) Replace(command string) string {
	parts := strings.Split(command, "|")

	out := []string{}

	for i, part := range parts {
		part = strings.TrimSpace(part)

		if i != len(parts)-1 {
			if replacement, ok := c.Mocks[part]; ok {
				part = replacement
			}
		}
		out = append(out, part)
	}

	return strings.Join(out, " | ")
}

func (c *MockConfiguration) ReplaceFiles(command string) string {
	out := command

	for filename, testFilename := range c.Files {
		out = strings.ReplaceAll(out, filename, testFilename)
	}
	return out
}
