package iostreams

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cli/safeexec"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/clio"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/vbauerster/mpb/v6"
	"golang.org/x/term"
)

type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer

	// the original (non-colorable) output stream
	originalOut   io.Writer
	colorEnabled  bool
	is256enabled  bool
	terminalTheme string

	progressIndicatorEnabled bool

	stdinTTYOverride  bool
	stdinIsTTY        bool
	stdoutTTYOverride bool
	stdoutIsTTY       bool
	stderrTTYOverride bool
	stderrIsTTY       bool

	Logger *logger.Logger

	neverPrompt bool

	TempFileOverride *os.File

	progress *mpb.Progress
}

func (s *IOStreams) SetProgress(v bool) {
	s.progressIndicatorEnabled = v
	s.progress = nil
}

func (s *IOStreams) SetColor(v bool) {
	s.colorEnabled = v
}

func (s *IOStreams) ColorEnabled() bool {
	return s.colorEnabled
}

func (s *IOStreams) ColorSupport256() bool {
	return s.is256enabled
}

func (s *IOStreams) DetectTerminalTheme() string {
	if !s.ColorEnabled() {
		s.terminalTheme = "none"
		return "none"
	}

	style := os.Getenv("GLAMOUR_STYLE")
	if style != "" && style != "auto" {
		s.terminalTheme = "none"
		return "none"
	}

	if termenv.HasDarkBackground() {
		s.terminalTheme = "dark"
		return "dark"
	}

	s.terminalTheme = "light"
	return "light"
}

func (s *IOStreams) TerminalTheme() string {
	if s.terminalTheme == "" {
		return "none"
	}

	return s.terminalTheme
}

func (s *IOStreams) SetStdinTTY(isTTY bool) {
	s.stdinTTYOverride = true
	s.stdinIsTTY = isTTY
}

func (s *IOStreams) IsStdinTTY() bool {
	if s.stdinTTYOverride {
		return s.stdinIsTTY
	}
	if stdin, ok := s.In.(*os.File); ok {
		return isTerminal(stdin)
	}
	return false
}

func (s *IOStreams) SetStdoutTTY(isTTY bool) {
	s.stdoutTTYOverride = true
	s.stdoutIsTTY = isTTY
}

func (s *IOStreams) IsStdoutTTY() bool {
	if s.stdoutTTYOverride {
		return s.stdoutIsTTY
	}
	if stdout, ok := s.Out.(*os.File); ok {
		return isTerminal(stdout)
	}
	return false
}

func (s *IOStreams) SetStderrTTY(isTTY bool) {
	s.stderrTTYOverride = true
	s.stderrIsTTY = isTTY
}

func (s *IOStreams) IsStderrTTY() bool {
	if s.stderrTTYOverride {
		return s.stderrIsTTY
	}
	if stderr, ok := s.ErrOut.(*os.File); ok {
		return isTerminal(stderr)
	}
	return false
}

func (s *IOStreams) HasStdin() bool {
	return clio.HasPipedInput(nil)
}

func (s *IOStreams) CanPrompt() bool {
	if s.neverPrompt {
		return false
	}

	return s.IsStdinTTY() && s.IsStdoutTTY()
}

func (s *IOStreams) CanPromptOnStdErr() bool {
	if s.neverPrompt {
		return false
	}

	return s.IsStdinTTY() && s.IsStderrTTY()
}

func (s *IOStreams) SetNeverPrompt(v bool) {
	s.neverPrompt = v
}

func (s *IOStreams) TerminalWidth() int {
	defaultWidth := 80
	out := s.Out
	if s.originalOut != nil {
		out = s.originalOut
	}

	if w, _, err := terminalSize(out); err == nil {
		return w
	}

	if isCygwinTerminal(out) {
		tputExe, err := safeexec.LookPath("tput")
		if err != nil {
			return defaultWidth
		}
		tputCmd := exec.Command(tputExe, "cols")
		tputCmd.Stdin = os.Stdin
		if out, err := tputCmd.Output(); err == nil {
			if w, err := strconv.Atoi(strings.TrimSpace(string(out))); err == nil {
				return w
			}
		}
	}

	return defaultWidth
}

func (s *IOStreams) ColorScheme() *ColorScheme {
	return NewColorScheme(s.ColorEnabled(), s.ColorSupport256())
}

func (s *IOStreams) ReadUserFile(fn string) ([]byte, error) {
	var r io.ReadCloser
	if fn == "-" {
		r = s.In
	} else {
		var err error
		r, err = os.Open(fn)
		if err != nil {
			return nil, err
		}
	}
	defer r.Close()
	return io.ReadAll(r)
}

func (s *IOStreams) TempFile(dir, pattern string) (*os.File, error) {
	if s.TempFileOverride != nil {
		return s.TempFileOverride, nil
	}
	return os.CreateTemp(dir, pattern)
}

func (s *IOStreams) ProgressIndicator() *mpb.Progress {
	if s.progressIndicatorEnabled {
		if s.progress == nil {
			s.progress = mpb.New(
				mpb.WithOutput(s.ErrOut),
				mpb.WithRefreshRate(180*time.Millisecond),
			)
		}
	}
	return s.progress
}

func (s *IOStreams) WaitForProgressIndicator() {
	if s.progress != nil {
		s.progress.Wait()
	}
}

// Environment variables to force the terminal (TTY) detection of the standard streams.
// These are intended for testing only, e.g. to show progress bars when stderr is not a terminal.
// Accepts boolean values, e.g. "true" or "false"
const (
	EnvForceStdinTTY  = "C8Y_FORCE_STDIN_TTY"
	EnvForceStdoutTTY = "C8Y_FORCE_STDOUT_TTY"
	EnvForceStderrTTY = "C8Y_FORCE_STDERR_TTY"
)

// forcedTTY returns the value of an environment variable used to force the terminal detection.
// The second return value is false if the variable is not set or is not a valid boolean
func forcedTTY(name string) (isTTY bool, ok bool) {
	value, found := os.LookupEnv(name)
	if !found {
		return false, false
	}
	isTTY, err := strconv.ParseBool(value)
	if err != nil {
		return false, false
	}
	return isTTY, true
}

func System(colorDisabled bool, colorForced bool) *IOStreams {
	stdoutIsTTY := isTerminal(os.Stdout)
	if v, ok := forcedTTY(EnvForceStdoutTTY); ok {
		stdoutIsTTY = v
	}
	stderrIsTTY := isTerminal(os.Stderr)
	if v, ok := forcedTTY(EnvForceStderrTTY); ok {
		stderrIsTTY = v
	}

	io := &IOStreams{
		In:           os.Stdin,
		originalOut:  os.Stdout,
		Out:          colorable.NewColorable(os.Stdout),
		ErrOut:       colorable.NewColorable(os.Stderr),
		colorEnabled: colorForced || (!colorDisabled && stdoutIsTTY),
		is256enabled: Is256ColorSupported(),
	}

	if stderrIsTTY {
		io.progressIndicatorEnabled = true
	}

	// prevent duplicate isTerminal queries now that we know the answer
	io.SetStdoutTTY(stdoutIsTTY)
	io.SetStderrTTY(stderrIsTTY)
	if v, ok := forcedTTY(EnvForceStdinTTY); ok {
		io.SetStdinTTY(v)
	}
	return io
}

func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return &IOStreams{
		In:     io.NopCloser(in),
		Out:    out,
		ErrOut: errOut,
	}, in, out, errOut
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

func isCygwinTerminal(w io.Writer) bool {
	if f, isFile := w.(*os.File); isFile {
		return isatty.IsCygwinTerminal(f.Fd())
	}
	return false
}

func terminalSize(w io.Writer) (int, int, error) {
	if f, isFile := w.(*os.File); isFile {
		return term.GetSize(int(f.Fd()))
	}
	return 0, 0, fmt.Errorf("%v is not a file", w)
}
