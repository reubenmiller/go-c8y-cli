package cmdutil

import (
	"github.com/reubenmiller/go-c8y-cli/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/console"
	"github.com/reubenmiller/go-c8y-cli/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

type Factory struct {
	IOStreams      *iostreams.IOStreams
	Client         func() (*c8y.Client, error)
	Config         func() (*config.Config, error)
	Logger         func() (*logger.Logger, error)
	ActivityLogger func() (*activitylogger.ActivityLogger, error)
	Console        func() (*console.Console, error)

	// Executable is the path to the currently invoked binary
	Executable string
}
