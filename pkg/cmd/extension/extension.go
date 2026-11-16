package extension

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/extensions"
	"gopkg.in/yaml.v3"
)

const manifestName = "manifest.yml"
const fileAlias = "extension.yaml"
const templateName = "templates"
const viewsName = "views"
const apiName = "api"
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
	isSystem       bool
	isPinned       bool
	currentVersion string
	latestVersion  string
	kind           ExtensionKind

	// manifest is shared between copies of the extension, so the file is only read once
	manifest *manifestCache
}

type manifestCache struct {
	once    sync.Once
	file    *ExtensionFile
	aliases []extensions.Alias
	err     error
}

// ExtensionFile is the extension manifest (extension.yaml)
type ExtensionFile struct {
	extensions.Manifest `yaml:",inline"`
	Aliases             []AliasExtension `json:"aliases,omitempty" yaml:"aliases,omitempty"`
}

// readManifest parses the extension.yaml file in the given extension folder. A missing file
// results in an empty manifest
func readManifest(dir string) (*ExtensionFile, error) {
	ext := &ExtensionFile{}
	path := filepath.Join(dir, fileAlias)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ext, nil
		}
		return ext, err
	}
	if err := yaml.Unmarshal(b, ext); err != nil {
		return &ExtensionFile{}, fmt.Errorf("invalid extension manifest. file=%s, reason=%w", path, err)
	}
	return ext, nil
}

func (e *Extension) loadManifest() *manifestCache {
	if e.manifest == nil {
		e.manifest = &manifestCache{}
	}
	e.manifest.once.Do(func() {
		file, err := readManifest(e.path)
		e.manifest.file = file
		e.manifest.err = err
		aliases := make([]extensions.Alias, 0)
		for i, alias := range file.Aliases {
			if alias.GetName() != "" && alias.GetCommand() != "" {
				aliases = append(aliases, &file.Aliases[i])
			}
		}
		e.manifest.aliases = aliases
	})
	return e.manifest
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
		e.isSystem ||
		e.currentVersion == "" ||
		e.latestVersion == "" ||
		e.currentVersion == e.latestVersion {
		return false
	}
	return true
}

// IsBinary returns true if the extension was installed from a binary release
func (e *Extension) IsBinary() bool {
	return e.kind == BinaryKind
}

// Kind returns how the extension was installed
func (e *Extension) Kind() extensions.Kind {
	if e.isSystem {
		return extensions.KindSystem
	}
	if e.isLocal {
		return extensions.KindLocal
	}
	if e.kind == BinaryKind {
		return extensions.KindBinary
	}
	return extensions.KindGit
}

// Manifest returns the parsed extension manifest (extension.yaml)
func (e *Extension) Manifest() (*extensions.Manifest, error) {
	m := e.loadManifest()
	return &m.file.Manifest, m.err
}

// Executable returns the absolute path to the binary declared in the extension manifest.
// It returns an empty string if no binary is declared
func (e *Extension) Executable() string {
	m := e.loadManifest()
	if m.file.Binary == nil {
		return ""
	}
	binPath := m.file.Binary.Path
	if binPath == "" {
		binPath = filepath.Join("bin", e.originalName())
	}
	return resolveExecutable(e.path, binPath, runtime.GOOS)
}

// originalName returns the name of the extension before it was installed (which can differ when
// installed with --name): the repository name for binary releases, and the source folder name
// for local extensions
func (e *Extension) originalName() string {
	if b, err := os.ReadFile(filepath.Join(e.path, manifestName)); err == nil {
		var bm binManifest
		if yaml.Unmarshal(b, &bm) == nil && bm.Name != "" {
			return bm.Name
		}
	}
	if resolved, err := filepath.EvalSymlinks(e.path); err == nil {
		return filepath.Base(resolved)
	}
	return filepath.Base(e.path)
}

// resolveExecutable returns the path of the binary within the extension folder.
// If binPath is empty, then bin/<folder-name> is used. On Windows, .exe is added if
// the path does not have a file extension
func resolveExecutable(dir string, binPath string, goos string) string {
	if binPath == "" {
		binPath = filepath.Join("bin", filepath.Base(dir))
	}
	binPath = filepath.FromSlash(binPath)
	if goos == "windows" && filepath.Ext(binPath) == "" {
		binPath += ".exe"
	}
	if filepath.IsAbs(binPath) {
		return binPath
	}
	return filepath.Join(dir, binPath)
}

// Custom extension components
func (e *Extension) Aliases() ([]extensions.Alias, error) {
	m := e.loadManifest()
	if m.err != nil {
		return nil, m.err
	}
	return m.aliases, nil
}

func (e *Extension) Commands() ([]extensions.Command, error) {
	path := filepath.Join(e.path, commandsName)
	commands := make([]extensions.Command, 0)

	err := filepath.WalkDir(path, func(ipath string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !strings.HasPrefix(info.Name(), ".") {
			commands = append(commands, &Command{
				name:    strings.TrimSpace(strings.ReplaceAll(strings.TrimPrefix(ipath, path), string(os.PathSeparator), " ")),
				command: ipath,
			})
		}
		return nil
	})

	// Ignore if the folder does not exist, as commands are not mandatory
	if os.IsNotExist(err) {
		return commands, nil
	}
	return commands, err
}

func (e *Extension) TemplatePath() string {
	return filepath.Join(e.path, templateName)
}

func (e *Extension) ViewPath() string {
	return filepath.Join(e.path, viewsName)
}
