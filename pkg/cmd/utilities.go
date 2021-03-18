package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
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

func getSessionHomeDir() string {
	var outputDir string

	if v := os.Getenv("C8Y_SESSION_HOME"); v != "" {
		expandedV, err := homedir.Expand(v)
		outputDir = v
		if err == nil {
			outputDir = expandedV
		} else {
			Logger.Warnf("Could not expand path. %s", err)
		}
	} else {
		outputDir, _ = homedir.Dir()
		outputDir = filepath.Join(outputDir, ".cumulocity", "sessions")
	}

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		Logger.Errorf("Could not create sessions directory and it does not exist. %s", err)
	}
	return outputDir
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

func showClientEnvironmentVariables(c8yclient *c8y.Client, shell ShellType) {
	output := cliConfig.GetEnvironmentVariables(c8yclient, false)
	showEnvironmentVariables(output, shell)
}

func showEnvironmentVariables(config map[string]interface{}, shell ShellType) {
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

// clearEnvironmentVariables clears all the session related environment variables by passing
// a shell snippet to execute via source or eval.
func clearEnvironmentVariables(shell ShellType) {
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

func checkEncryption(w io.Writer) error {
	encryptionEnabled := cliConfig.IsEncryptionEnabled()
	decryptSession := false
	if !encryptionEnabled && cliConfig.IsPasswordEncrypted() {
		cliConfig.Logger.Infof("Encryption has been disabled but detected a encrypted session")
		decryptSession = true
	}
	if encryptionEnabled || cliConfig.IsPasswordEncrypted() {
		if err := cliConfig.ReadKeyFile(); err != nil {
			return err
		}

		// check if encryption is used on the current session
		passphrase, err := cliConfig.CheckEncryption()
		if err != nil {
			return err
		}
		if passphrase == "" || passphrase == "null" {
			return fmt.Errorf("passphrase can not be empty")
		}
		cliConfig.Passphrase = passphrase

		// Decrypt username and cookies if necessary
		clientpass, err := cliConfig.GetPassword()
		if err != nil {
			return err
		}
		client.SetCookies(cliConfig.GetCookies())
		client.Password = clientpass

		// decrypt settings
		if decryptSession {
			if err := cliConfig.DecryptSession(); err != nil {
				return err
			}
		}

		green := promptui.Styler(promptui.FGGreen)
		fmt.Fprint(w, green("Passphrase OK\n"))
		Logger.Info("Passphrase accepted")
	}
	return nil
}
