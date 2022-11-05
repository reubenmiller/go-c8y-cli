package extensions

import (
	"io"

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
	Path() string // Path to executable
	URL() string
	CurrentVersion() string
	IsPinned() bool
	UpdateAvailable() bool
	IsBinary() bool
	IsLocal() bool

	// Extension components
	TemplatePath() string
	ViewPath() string
	Aliases() ([]Alias, error)
	Commands() ([]Command, error)
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

//go:generate moq -rm -out manager_mock.go . ExtensionManager
type ExtensionManager interface {
	List() []Extension
	Install(ghrepo.Interface, string, string) error
	InstallLocal(dir string, name string) error
	Upgrade(name string, force bool) error
	Remove(name string) error
	Dispatch(args []string, stdin io.Reader, stdout, stderr io.Writer) (bool, error)
	Create(name string, tmplType ExtTemplateType) error
	EnableDryRunMode()
}
