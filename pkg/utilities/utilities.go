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

	"github.com/manifoldco/promptui"
	"github.com/reubenmiller/go-c8y-cli/pkg/config"
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

	// Use the net/http package's handy DectectContentType function. Always returns a valid
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

func ShowClientEnvironmentVariables(cfg *config.Config, c8yclient *c8y.Client, shell ShellType) {
	output := cfg.GetEnvironmentVariables(c8yclient, false)
	ShowEnvironmentVariables(output, shell)
}

func ShowEnvironmentVariables(config map[string]interface{}, shell ShellType) {
	// sort output variables
	variables := []string{}

	for name := range config {
		variables = append(variables, name)
	}
	sort.Strings(variables)
	for _, name := range variables {
		value := config[name]

		switch shell {
		case ShellPowerShell:
			fmt.Printf("$env:%s = '%v'\n", name, value)
		case ShellFish:
			fmt.Printf("set -gx %s '%v'\n", name, value)
		default:
			fmt.Printf("export %s='%v'\n", name, value)
		}
	}
}

// ClearEnvironmentVariables clears all the session related environment variables by passing
// a shell snippet to execute via source or eval.
func ClearEnvironmentVariables(shell ShellType) {
	variables := []string{
		"C8Y_HOST",
		"C8Y_URL",
		"C8Y_BASEURL",
		"C8Y_TENANT",
		"C8Y_USER",
		"C8Y_USERNAME",
		"C8Y_PASSWORD",
		"C8Y_SESSION",
		"C8Y_SETTINGS_MODE_ENABLECREATE",
		"C8Y_SETTINGS_MODE_ENABLEUPDATE",
		"C8Y_SETTINGS_MODE_ENABLEDELETE",
		"C8Y_CREDENTIAL_COOKIES_0",
		"C8Y_CREDENTIAL_COOKIES_1",
		"C8Y_CREDENTIAL_COOKIES_2",
		"C8Y_CREDENTIAL_COOKIES_3",
		"C8Y_CREDENTIAL_COOKIES_4",
	}

	sort.Strings(variables)
	for _, name := range variables {
		switch shell {
		case ShellPowerShell:
			fmt.Printf("$env:%s = $null\n", name)
		case ShellFish:
			fmt.Printf("set -u %s\n", name)
		default:
			fmt.Printf("export %s=\n", name)
		}
	}
}

func CheckEncryption(w io.Writer, cfg *config.Config, client *c8y.Client) error {
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

		// Decrypt username and cookies if necessary
		clientpass, err := cfg.GetPassword()
		if err != nil {
			return err
		}
		client.SetCookies(cfg.GetCookies())
		client.Password = clientpass

		// decrypt settings
		if decryptSession {
			if err := cfg.DecryptSession(); err != nil {
				return err
			}
		}

		green := promptui.Styler(promptui.FGGreen)
		fmt.Fprint(w, green("Passphrase OK\n"))
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