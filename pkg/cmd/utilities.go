package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"

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

// saveResponseToFile saves a response to file
// @filename	filename
// @directory	output directory. If empty, then a temp directory will be used
// if filename
func saveResponseToFile(resp *c8y.Response, filename string, append bool, newline bool) (string, error) {

	var out *os.File
	var err error
	if append {
		out, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		out, err = os.Create(filename)
	}

	if err != nil {
		return "", fmt.Errorf("Could not create file. %s", err)
	}
	defer out.Close()

	// Writer the body to file
	Logger.Printf("header: %v", resp.Header)
	_, err = io.Copy(out, resp.Body)

	if newline {
		// add trailing newline so that json lines are seperated by lines
		fmt.Fprintf(out, "\n")
	}
	if err != nil {
		return "", fmt.Errorf("failed to copy file contents to file. %s", err)
	}

	if fullpath, err := filepath.Abs(filename); err == nil {
		return fullpath, nil
	}
	return filename, nil
}

func getSessionHomeDir() string {
	var outputDir string

	if v := os.Getenv("C8Y_SESSION_HOME"); v != "" {
		outputDir = v
	} else {
		outputDir, _ = homedir.Dir()
		outputDir = filepath.Join(outputDir, ".cumulocity")
	}
	return outputDir
}

func showEnvironmentVariables(c8yclient *c8y.Client, isPowerShell bool) {
	// sort output variables
	variables := []string{}
	output := cliConfig.GetEnvironmentVariables(c8yclient, false)
	for name := range output {
		variables = append(variables, name)
	}
	sort.Strings(variables)
	for _, name := range variables {
		value := output[name]

		if isPowerShell {
			fmt.Printf("$env:%s = '%s'\n", name, value)
		} else {
			fmt.Printf("export %s='%s'\n", name, value)
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
