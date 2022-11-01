package extension

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
)

const manifestName = "manifest.yml"
const fileAlias = "extension.json"
const templateName = "templates"
const viewsName = "views"
const commandsName = "commands"

type ExtensionKind int

const (
	GitKind ExtensionKind = iota
	BinaryKind
)

type Extension struct {
	path           string
	url            string
	isLocal        bool
	isPinned       bool
	currentVersion string
	latestVersion  string
	kind           ExtensionKind

	aliases []extensions.Alias
}

type ExtensionFile struct {
	Aliases []AliasExtension `json:"aliases,omitempty"`
}

func (e *Extension) Name() string {
	return strings.TrimPrefix(filepath.Base(e.path), ExtPrefix)
}

func (e *Extension) Path() string {
	return e.path
}

func (e *Extension) URL() string {
	return e.url
}

func (e *Extension) IsLocal() bool {
	return e.isLocal
}

func (e *Extension) CurrentVersion() string {
	return e.currentVersion
}

func (e *Extension) IsPinned() bool {
	return e.isPinned
}

func (e *Extension) UpdateAvailable() bool {
	if e.isPinned ||
		e.isLocal ||
		e.currentVersion == "" ||
		e.latestVersion == "" ||
		e.currentVersion == e.latestVersion {
		return false
	}
	return true
}

func (e *Extension) IsBinary() bool {
	return e.kind == BinaryKind
}

// Custom extension components
func (e *Extension) Aliases() ([]extensions.Alias, error) {
	if len(e.aliases) > 0 {
		return e.aliases, nil
	}
	path := filepath.Join(e.path, fileAlias)
	aliases := make([]extensions.Alias, 0)

	if file, err := os.Open(path); err == nil {
		if b, bErr := io.ReadAll(file); bErr == nil {
			ext := &ExtensionFile{}
			if jErr := json.Unmarshal(b, ext); jErr != nil {
				return nil, jErr
			}

			for i, alias := range ext.Aliases {
				if alias.GetName() != "" && alias.GetCommand() != "" {
					aliases = append(aliases, &ext.Aliases[i])
				}
			}
		}
	}
	e.aliases = aliases
	return aliases, nil
}

func (e *Extension) Commands() ([]extensions.Command, error) {
	path := filepath.Join(e.path, commandsName)
	commands := make([]extensions.Command, 0)

	dirs, err := os.ReadDir(path)
	if err != nil {
		return commands, err
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			commands = append(commands, &Command{
				name:    filepath.Base(dir.Name()),
				command: dir.Name(),
			})
		}
	}

	return commands, nil
}

func (e *Extension) TemplatePath() string {
	return filepath.Join(e.path, templateName)
}

func (e *Extension) ViewPath() string {
	return filepath.Join(e.path, viewsName)
}
