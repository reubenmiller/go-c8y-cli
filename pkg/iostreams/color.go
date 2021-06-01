package iostreams

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	magenta  = color.New(color.FgMagenta)
	cyan     = color.New(color.FgCyan)
	red      = color.New(color.FgRed)
	yellow   = color.New(color.FgYellow)
	blue     = color.New(color.FgBlue)
	green    = color.New(color.FgGreen)
	gray     = color.New(color.BgHiBlack)
	bold     = color.New(color.Bold)
	cyanBold = color.New(color.FgCyan, color.Bold)

	gray256 = func(t string) string {
		return fmt.Sprintf("\x1b[%d;5;%dm%s\x1b[m", 38, 242, t)
	}
)

func Is256ColorSupported() bool {
	term := os.Getenv("TERM")
	colorterm := os.Getenv("COLORTERM")

	return strings.Contains(term, "256") ||
		strings.Contains(term, "24bit") ||
		strings.Contains(term, "truecolor") ||
		strings.Contains(colorterm, "256") ||
		strings.Contains(colorterm, "24bit") ||
		strings.Contains(colorterm, "truecolor")
}

func NewColorScheme(enabled, is256enabled bool) *ColorScheme {
	return &ColorScheme{
		enabled:      enabled,
		is256enabled: is256enabled,
	}
}

type ColorScheme struct {
	enabled      bool
	is256enabled bool
}

func (c *ColorScheme) Bold(t string) string {
	if !c.enabled {
		return t
	}
	return bold.Sprint(t)
}

func (c *ColorScheme) Red(t string) string {
	if !c.enabled {
		return t
	}
	return red.Sprint(t)
}

func (c *ColorScheme) Yellow(t string) string {
	if !c.enabled {
		return t
	}
	return yellow.Sprint(t)
}

func (c *ColorScheme) Green(t string) string {
	if !c.enabled {
		return t
	}
	return green.Sprint(t)
}

func (c *ColorScheme) Gray(t string) string {
	if !c.enabled {
		return t
	}
	if c.is256enabled {
		return gray256(t)
	}
	return gray.Sprint(t)
}

func (c *ColorScheme) Magenta(t string) string {
	if !c.enabled {
		return t
	}
	return magenta.Sprint(t)
}

func (c *ColorScheme) Cyan(t string) string {
	if !c.enabled {
		return t
	}
	return cyan.Sprint(t)
}

func (c *ColorScheme) CyanBold(t string) string {
	if !c.enabled {
		return t
	}
	return cyanBold.Sprint(t)
}

func (c *ColorScheme) Blue(t string) string {
	if !c.enabled {
		return t
	}
	return blue.Sprint(t)
}

func (c *ColorScheme) SuccessIcon() string {
	return c.SuccessIconWithColor(c.Green)
}

func (c *ColorScheme) SuccessIconWithColor(colo func(string) string) string {
	return colo("✓")
}

func (c *ColorScheme) WarningIcon() string {
	return c.Yellow("!")
}

func (c *ColorScheme) FailureIcon() string {
	return c.Red("X")
}

func (c *ColorScheme) ColorFromString(s string) func(string) string {
	s = strings.ToLower(s)
	var fn func(string) string
	switch s {
	case "bold":
		fn = c.Bold
	case "red":
		fn = c.Red
	case "yellow":
		fn = c.Yellow
	case "green":
		fn = c.Green
	case "gray":
		fn = c.Gray
	case "magenta":
		fn = c.Magenta
	case "cyan":
		fn = c.Cyan
	case "blue":
		fn = c.Blue
	default:
		fn = func(s string) string {
			return s
		}
	}

	return fn
}
