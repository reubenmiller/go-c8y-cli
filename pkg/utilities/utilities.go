package utilities

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

// GetFileContentType TODO: Fix mime detection because it currently returns only application/octet-stream
func GetFileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DetectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

type ShellType int

func (t ShellType) FromString(name string) ShellType {
	values := map[string]ShellType{
		"powershell": ShellPowerShell,
		"bash":       ShellBash,
		"zsh":        ShellZSH,
		"fish":       ShellFish,
	}

	if v, ok := values[strings.ToLower(name)]; ok {
		return v
	}
	return t
}

func (t ShellType) Parse(name string) (ShellType, bool) {
	values := map[string]ShellType{
		"powershell": ShellPowerShell,
		"bash":       ShellBash,
		"zsh":        ShellZSH,
		"fish":       ShellFish,
	}

	v, ok := values[strings.ToLower(name)]
	return v, ok
}

const (
	// ShellBash bash
	ShellBash ShellType = iota

	// ShellZSH zsh
	ShellZSH

	// ShellPowerShell PowerShell
	ShellPowerShell

	// ShellFish fish
	ShellFish
)

func WriteShellVariables(w io.Writer, cfg map[string]interface{}, shell ShellType) {
	// sort output variables
	variables := []string{}

	for name := range cfg {
		variables = append(variables, name)
	}
	sort.Strings(variables)
	for _, name := range variables {
		value := cfg[name]

		switch shell {
		case ShellPowerShell:
			if value == "" {
				fmt.Fprintf(w, "$env:%s = $null\n", name)
			} else {
				fmt.Fprintf(w, "$env:%s = '%v'\n", name, value)
			}
		case ShellFish:
			if value == "" {
				fmt.Fprintf(w, "set -e %s\n", name)
			} else {
				fmt.Fprintf(w, "set -gx %s '%v'\n", name, value)
			}
		default:
			if value == "" {
				fmt.Fprintf(w, "unset %s\n", name)
			} else {
				fmt.Fprintf(w, "export %s='%v'\n", name, value)
			}
		}
	}
}

// ClearEnvironmentVariables clears all the session related environment variables by passing
// a shell snippet to execute via source or eval.
func ClearEnvironmentVariables(variables []string, shell ShellType) {
	sort.Strings(variables)
	for _, name := range variables {
		switch shell {
		case ShellPowerShell:
			fmt.Printf("$env:%s = $null\n", name)
		case ShellFish:
			fmt.Printf("set -e %s\n", name)
		default:
			fmt.Printf("unset %s\n", name)
		}
	}
}

func ClearProcessEnvironment(variables []string) {
	for _, key := range variables {
		os.Unsetenv(key)
	}
}

func CheckEncryption(IO *iostreams.IOStreams, cfg *config.Config, client *c8y.Client) error {
	encryptionEnabled := cfg.IsEncryptionEnabled()
	decryptSession := false
	if !encryptionEnabled && cfg.IsPasswordEncrypted() {
		cfg.Logger.Infof("Encryption has been disabled but detected a encrypted session")
		decryptSession = true
	}
	if encryptionEnabled || cfg.IsPasswordEncrypted() {
		if err := cfg.ReadKeyFile(); err != nil {
			return err
		}

		// check if encryption is used on the current session
		passphrase, err := cfg.CheckEncryption()
		if err != nil {
			return err
		}
		if passphrase == "" || passphrase == "null" {
			return fmt.Errorf("passphrase can not be empty")
		}
		cfg.Passphrase = passphrase

		// Decrypt username and token if necessary
		clientpass, err := cfg.GetPassword()
		if err != nil {
			return err
		}
		client.Password = clientpass
		clienttoken, err := cfg.GetToken()
		if err != nil {
			return err
		}
		client.SetToken(clienttoken)

		// decrypt settings
		if decryptSession {
			if err := cfg.DecryptSession(); err != nil {
				return err
			}
		}

		cs := IO.ColorScheme()
		successMsg := fmt.Sprintf("%s Passphrase OK\n", cs.SuccessIcon())
		fmt.Fprint(IO.ErrOut, successMsg)
	}
	return nil
}

func GetCommandLineArgs() string {
	buf := bytes.Buffer{}

	escapeQuotesWindows := func(value string) string {
		return "'" + strings.ReplaceAll(value, "\"", `\"`) + "'"
	}
	escapeQuotesLinux := func(value string) string {
		return "\"" + strings.ReplaceAll(value, "\"", `\"`) + "\""
	}

	quoter := escapeQuotesLinux
	if runtime.GOOS == "windows" {
		quoter = escapeQuotesWindows
	}

	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, " ") {
			if strings.HasPrefix(arg, "-") {
				if index := strings.Index(arg, "="); index > -1 {
					buf.WriteString(arg[0:index] + "=")
					buf.WriteString(quoter(arg[index+1:]))
				} else {
					buf.WriteString(quoter(arg))
				}
			} else {
				buf.WriteString(quoter(arg))
			}
		} else {
			buf.WriteString(arg)
		}
		buf.WriteByte(' ')
	}
	return buf.String()
}
