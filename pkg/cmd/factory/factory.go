package factory

import (
	"os"

	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

func New(appVersion string, configFunc func() (*config.Config, error), clientFunc func() (*c8y.Client, error), loggerFunc func() (*logger.Logger, error)) *cmdutil.Factory {
	io := iostreams.System(false, true)

	c8yExecutable := "c8y"
	if exe, err := os.Executable(); err == nil {
		c8yExecutable = exe
	}

	return &cmdutil.Factory{
		IOStreams:  io,
		Config:     configFunc,
		Client:     clientFunc,
		Executable: c8yExecutable,
		Logger:     loggerFunc,
	}
}
