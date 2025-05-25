package extensions_wasi

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

type RuntimeWASI struct {
	Module  api.Module
	Runtime wazero.Runtime
}

type RuntimeOptions struct {
	IO      *iostreams.IOStreams
	File    string
	Args    []string
	Logger  *logger.Logger
	Timeout time.Duration
}

func NewRuntimeWASI(ctx context.Context, options RuntimeOptions) error {
	// Create a new WebAssembly Runtime.
	rtc := wazero.NewRuntimeConfig()

	if options.Timeout > 0 {
		newCtx, cancel := context.WithTimeout(ctx, options.Timeout)
		options.Logger.Debugf("Setting runtime timeout to %v", options.Timeout)
		ctx = newCtx
		defer cancel()
		rtc = rtc.WithCloseOnContextDone(true)
	} else if options.Timeout < 0 {
		return fmt.Errorf("timeout duration may not be negative, %v given", options.Timeout)
	}

	rt := wazero.NewRuntimeWithConfig(ctx, rtc)
	defer rt.Close(ctx) // This closes everything this Runtime created.

	// Instantiate a Go-defined module named "env" that exports functions to
	// get the current year and log to the console.
	//
	// Note: As noted on wazero.HostFunctionBuilder documentation, function
	// signatures are constrained to a subset of numeric types.
	// Note: "env" is a module name conventionally used for arbitrary
	// host-defined functions, but any name would do.
	_, err := rt.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(func(v uint32) {
			fmt.Println("log_i32 >>", v)
		}).
		Export("check").
		NewFunctionBuilder().
		WithFunc(func() uint32 {
			if envYear, err := strconv.ParseUint(os.Getenv("CURRENT_YEAR"), 10, 64); err == nil {
				return uint32(envYear) // Allow env-override to prevent annual test maintenance!
			}
			return uint32(time.Now().Year())
		}).
		Export("current_year").
		Instantiate(ctx)
	if err != nil {
		log.Panicln(err)
	}

	wasm, readErr := os.ReadFile(options.File)
	if readErr != nil {
		return readErr
	}

	stdinBuf := bytes.NewBufferString("test input")

	stdOutBuf := bytes.NewBufferString("")
	stdOut := io.MultiWriter(os.Stdout, stdOutBuf)

	stdErr := os.Stderr

	conf := wazero.NewModuleConfig().
		WithStartFunctions("_start", "_initialize").
		WithArgs(options.Args...).
		WithStdout(stdOut).
		WithStderr(stdErr).
		WithStdin(stdinBuf).
		WithRandSource(rand.Reader).
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime()

	for _, item := range os.Environ() {
		if strings.HasPrefix(item, "C8Y_SETTINGS") || strings.HasPrefix(item, "C8Y_TENANT") {
			if key, value, found := strings.Cut(item, "="); found {
				options.Logger.Debugf("Adding env. %s=%s", key, value)
				conf = conf.WithEnv(key, value)
			}
		}
	}

	guest, err := rt.CompileModule(ctx, wasm)
	if err != nil {
		return fmt.Errorf("error compiling wasm binary: %w", err)
	}

	var mod api.Module

	switch detectImports(guest.ImportedFunctions()) {
	case modeWasi:
		wasi_snapshot_preview1.MustInstantiate(ctx, rt)
		mod, err = rt.InstantiateModule(ctx, guest, conf)
	case modeWasiUnstable:
		// Instantiate the current WASI functions under the wasi_unstable
		// instead of wasi_snapshot_preview1.
		wasiBuilder := rt.NewHostModuleBuilder("wasi_unstable")
		wasi_snapshot_preview1.NewFunctionExporter().ExportFunctions(wasiBuilder)
		mod, err = wasiBuilder.Instantiate(ctx)
		if err == nil {
			// Instantiate our binary, but using the old import names.
			mod, err = rt.InstantiateModule(ctx, guest, conf)
		}
	case modeDefault:
		mod, err = rt.InstantiateModule(ctx, guest, conf)
	}

	if err != nil {
		if exitErr, ok := err.(*sys.ExitError); ok {
			exitCode := exitErr.ExitCode()
			if exitCode == sys.ExitCodeDeadlineExceeded {
				return cmderrors.NewErrorWithExitCode(cmderrors.ExitTimeout, fmt.Errorf("error: %v (timeout %v)", exitErr, options.Timeout))
			}
			return cmderrors.NewErrorWithExitCode(cmderrors.ExitCode(exitCode), err)
		}
		return cmderrors.NewErrorWithExitCode(cmderrors.ExitCode(1), fmt.Errorf("error instantiating wasm binary: %w", err))
	}

	commandName := ""
	if len(options.Args) > 1 {
		commandName = options.Args[1]
	}
	options.Logger.Debugf("wasm arguments: %v", options.Args)

	// Calling function from inside
	for name := range mod.ExportedFunctionDefinitions() {
		options.Logger.Debugf("Exported function: %s", name)

		if name == commandName {
			options.Logger.Debugf("Calling function: %s", name)
			moduleFunc := mod.ExportedFunction(name)
			if moduleFunc == nil {
				options.Logger.Error("Exported function is nil")
				return fmt.Errorf("WASM module is broken")
			}
			_, err := moduleFunc.Call(ctx)
			if err != nil {
				return err
			}
		}
	}

	options.Logger.Debugf("------------------------\n\nExtension (stdout)\n%s", stdOutBuf.Bytes())

	return nil
}

const (
	modeDefault importMode = iota
	modeWasi
	modeWasiUnstable
)

type importMode uint

func detectImports(imports []api.FunctionDefinition) importMode {
	for _, f := range imports {
		moduleName, _, _ := f.Import()
		switch moduleName {
		case wasi_snapshot_preview1.ModuleName:
			return modeWasi
		case "wasi_unstable":
			return modeWasiUnstable
		}
	}
	return modeDefault
}
