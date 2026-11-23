package extensions

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/ghrepo"
)

type ExtTemplateType int

const (
	GitTemplateType      ExtTemplateType = 0
	GoBinTemplateType    ExtTemplateType = 1
	OtherBinTemplateType ExtTemplateType = 2
)

//go:generate moq -rm -out extension_mock.go . Extension
type Extension interface {
	Name() string // Extension Name without c8y-
	Path() string // Path to the extension folder
	URL() string
	CurrentVersion() string
	IsPinned() bool
	UpdateAvailable() bool
	IsBinary() bool
	IsLocal() bool

	// Kind is how the extension was installed (git, local, binary or system)
	Kind() Kind

	// Executable returns the absolute path to the extension's binary, as declared in the
	// extension manifest (extension.yaml). It is empty if the extension does not declare a binary.
	// The file is not checked for existence.
	Executable() string

	// Manifest returns the parsed extension manifest (extension.yaml). An empty manifest
	// is returned if the extension does not have one
	Manifest() (*Manifest, error)

	// Extension components
	TemplatePath() string
	ViewPath() string
	Aliases() ([]Alias, error)
	Commands() ([]Command, error)
}

// Kind is how an extension was installed
type Kind string

const (
	KindGit    Kind = "git"
	KindLocal  Kind = "local"
	KindBinary Kind = "binary"
	// KindSystem is an extension in a system-wide extensions folder, installed by a package manager
	KindSystem Kind = "system"
)

// Manifest is the extension manifest file (extension.yaml)
type Manifest struct {
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// Name is the extension name (without the c8y- prefix), which is used when the extension is
	// installed, unless the user passes --name. See NormalizeName
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Binary declares an executable provided by the extension
	Binary *BinarySpec `json:"binary,omitempty" yaml:"binary,omitempty"`

	// Commands declares commands which are provided by the binary
	Commands []BinaryCommand `json:"commands,omitempty" yaml:"commands,omitempty"`

	// SessionProviders declares session providers which are provided by the binary
	SessionProviders []SessionProvider `json:"sessionProviders,omitempty" yaml:"sessionProviders,omitempty"`
}

// Session provider output formats
const (
	ProviderFormatAuto   = "auto"
	ProviderFormatJSON   = "json"
	ProviderFormatDotenv = "dotenv"
)

// SessionProvider is a session provider which is run by "c8y sessions login" to select a session
type SessionProvider struct {
	// Name of the provider
	Name string `json:"name" yaml:"name"`

	// Description shown in the completion
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Exec are the arguments passed to the binary before the provider flags
	Exec []string `json:"exec,omitempty" yaml:"exec,omitempty"`

	// Secrets are the names of environment variables which the provider needs
	Secrets []string `json:"secrets,omitempty" yaml:"secrets,omitempty"`

	// Format of the provider output: json, dotenv or auto (default)
	Format string `json:"format,omitempty" yaml:"format,omitempty"`

	// List are the arguments which make the binary print the provider's sessions as a json
	// array, without secrets and without user interaction. It is used to show the sessions in
	// other pickers, e.g. the browser session picker of "c8y mcp serve". Requires Select
	List []string `json:"list,omitempty" yaml:"list,omitempty"`

	// Select are the arguments which make the binary print one session, given by its
	// sessionUri (which is appended to the arguments), without user interaction. Requires List
	Select []string `json:"select,omitempty" yaml:"select,omitempty"`
}

// CanList returns true if the provider can list its sessions and select one by its sessionUri
// without user interaction
func (p *SessionProvider) CanList() bool {
	return len(p.List) > 0 && len(p.Select) > 0
}

// OutputFormat returns the declared output format, or an empty string for automatic detection
func (p *SessionProvider) OutputFormat() string {
	if strings.EqualFold(p.Format, ProviderFormatAuto) {
		return ""
	}
	return strings.ToLower(p.Format)
}

var sessionProviderName = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// FileSessionProvider is a reserved session provider name, which refers to the session files in
// the session home folder (the same name as the "file" session provider type), e.g. in
// "c8y mcp serve --sessionProvider file"
const FileSessionProvider = "file"

// ValidSessionProviders returns the usable session providers of the manifest, and warnings for
// providers which are ignored: all providers if no binary is declared, providers with an invalid
// name or format, and duplicate names (the first one is used)
func (m *Manifest) ValidSessionProviders() ([]SessionProvider, []string) {
	if len(m.SessionProviders) == 0 {
		return nil, nil
	}
	if m.Binary == nil {
		return nil, []string{"session providers are ignored as the extension does not declare a binary"}
	}
	providers := make([]SessionProvider, 0, len(m.SessionProviders))
	warnings := []string{}
	seen := map[string]bool{}
	for _, provider := range m.SessionProviders {
		switch {
		case !sessionProviderName.MatchString(provider.Name):
			warnings = append(warnings, fmt.Sprintf("session provider is ignored as the name is invalid. name=%q, expected=%s", provider.Name, sessionProviderName.String()))
		case provider.Name == FileSessionProvider:
			warnings = append(warnings, fmt.Sprintf("session provider is ignored as the name is reserved for the session files. name=%s", provider.Name))
		case seen[provider.Name]:
			warnings = append(warnings, fmt.Sprintf("duplicate session provider is ignored. name=%s", provider.Name))
		case provider.Format != "" && provider.OutputFormat() != "" && provider.OutputFormat() != ProviderFormatJSON && provider.OutputFormat() != ProviderFormatDotenv:
			warnings = append(warnings, fmt.Sprintf("session provider is ignored as the format is invalid. name=%s, format=%s", provider.Name, provider.Format))
		default:
			if (len(provider.List) > 0) != (len(provider.Select) > 0) {
				// the provider still works interactively
				warnings = append(warnings, fmt.Sprintf("session provider can't list sessions as only one of list and select is set. name=%s", provider.Name))
				provider.List = nil
				provider.Select = nil
			}
			seen[provider.Name] = true
			providers = append(providers, provider)
		}
	}
	return providers, warnings
}

// Binary completion modes
const (
	CompletionNone  = "none"
	CompletionCobra = "cobra"
)

// Binary command output modes
const (
	OutputRaw  = "raw"
	OutputJSON = "json"
)

// BinarySpec declares the binary provided by an extension
type BinarySpec struct {
	// Path to the binary, relative to the extension folder. Defaults to bin/<extension-dir-name>
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// Mount is the space separated command path (under the extension name) where all
	// unknown arguments are passed to the binary. Defaults to the extension root
	Mount string `json:"mount,omitempty" yaml:"mount,omitempty"`

	// Completion mode: cobra or none (default)
	Completion string `json:"completion,omitempty" yaml:"completion,omitempty"`
}

// MountPath returns the mount as a list of command names
func (b *BinarySpec) MountPath() []string {
	return strings.Fields(b.Mount)
}

// CompletionEnabled returns true if completion requests should be forwarded to the binary
func (b *BinarySpec) CompletionEnabled() bool {
	return strings.EqualFold(b.Completion, CompletionCobra)
}

// BinaryCommand is a command which is provided by the extension's binary
type BinaryCommand struct {
	// Name is the space separated command path under the extension name, e.g. "services list"
	Name string `json:"name" yaml:"name"`

	// Description shown in the help
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Exec are the arguments passed to the binary before the user's arguments.
	// Defaults to the name split on spaces
	Exec []string `json:"exec,omitempty" yaml:"exec,omitempty"`

	// Output mode: raw (default) or json
	Output string `json:"output,omitempty" yaml:"output,omitempty"`
}

// Path returns the command path as a list of command names
func (c *BinaryCommand) Path() []string {
	return strings.Fields(c.Name)
}

// ExecArgs returns the arguments which are passed to the binary before the user's arguments
func (c *BinaryCommand) ExecArgs() []string {
	if c.Exec != nil {
		return c.Exec
	}
	return c.Path()
}

// IsJSON returns true if the command output should be processed by c8y
func (c *BinaryCommand) IsJSON() bool {
	return strings.EqualFold(c.Output, OutputJSON)
}

//go:generate moq -rm -out alias_mock.go . Alias
type Alias interface {
	GetCommand() string
	GetName() string
	GetDescription() string
	IsShell() bool
}

type Command interface {
	Command() string
	Name() string
	Description() string
}

type Template interface {
	Path() string
	Name() string
}

type View interface {
	Path() string
	Name() string
}

// ExecOptions are the options to run an extension command
type ExecOptions struct {
	// Executable to run
	Executable string

	// Args passed to the executable
	Args []string

	// IsBinary runs the executable directly on all platforms. Otherwise it is
	// treated as a script, which is run via sh on Windows to support shebang lines
	IsBinary bool

	// Env is the environment of the process. The current environment is used if nil
	Env []string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// InstallOptions are the options of an extension install
type InstallOptions struct {
	// Name set by the user (--name). It takes precedence over the name declared in
	// extension.yaml and the repository or folder name
	Name string

	// Pin is the release tag or commit to install (remote installs only)
	Pin string

	// CheckName is called with the resolved name (without the c8y- prefix) before the
	// extension is moved into place. The install fails if it returns an error
	CheckName func(name string) error
}

//go:generate moq -rm -out manager_mock.go . ExtensionManager
type ExtensionManager interface {
	List() []Extension

	// Install installs an extension from a repository, and returns the name it was installed as
	Install(repo ghrepo.Interface, opts InstallOptions) (string, error)

	// InstallLocal installs an extension from a local folder, and returns the name it was installed as
	InstallLocal(dir string, opts InstallOptions) (string, error)
	Upgrade(name string, force bool) error
	Remove(name string) error
	Execute(opts ExecOptions) error
	Create(name string, tmplType ExtTemplateType) error
	EnableDryRunMode()
}
