package extensions

import (
	"io"
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

	// Binary declares an executable provided by the extension
	Binary *BinarySpec `json:"binary,omitempty" yaml:"binary,omitempty"`

	// Commands declares commands which are provided by the binary
	Commands []BinaryCommand `json:"commands,omitempty" yaml:"commands,omitempty"`
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

//go:generate moq -rm -out manager_mock.go . ExtensionManager
type ExtensionManager interface {
	List() []Extension
	Install(ghrepo.Interface, string, string) error
	InstallLocal(dir string, name string) error
	Upgrade(name string, force bool) error
	Remove(name string) error
	Execute(opts ExecOptions) error
	Create(name string, tmplType ExtTemplateType) error
	EnableDryRunMode()
}
