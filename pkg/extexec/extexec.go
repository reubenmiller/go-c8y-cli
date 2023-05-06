package extexec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/cli/safeexec"
)

func IsC8YCommand(cmd []string) bool {
	return len(cmd) > 0 && cmd[0] == "c8y"
}

func ExecuteExternalCommand(name string, fixed_args []string, optional_args ...string) ([]byte, error) {
	args := []string{}
	for _, a := range fixed_args {
		if strings.Contains(a, "%") {
			args = append(args, fmt.Sprintf(a, name))
		} else {
			args = append(args, a)
		}
	}
	if IsC8YCommand(args) {
		args = append(args, optional_args...)
	}

	exePath := ""
	var err error
	if len(args) < 1 {
		return nil, errors.New("invalid external command")
	}

	if args[0] == "c8y" {
		if runtime.GOOS == "windows" {
			exePath, err = safeexec.LookPath("c8y.exe")
			if err != nil {
				return nil, err
			}
		} else {
			exePath, err = safeexec.LookPath("c8y")
			if err != nil {
				return nil, err
			}
		}

		if ep, err := os.Executable(); err != nil {
			exePath = ep
		}
	} else {
		exePath = args[0]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	binaryCommand := exec.CommandContext(
		ctx,
		exePath,
		args[1:]...,
	)

	binaryCommand.Env = append(os.Environ(), "C8Y_SETTINGS_DEFAULTS_OUTPUT=csv", "C8Y_SETTINGS_DEFAULTS_PAGESIZE=20")
	binaryCommand.Dir = "."
	return binaryCommand.Output()
}
