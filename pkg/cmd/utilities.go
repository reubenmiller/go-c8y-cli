package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"

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

// getTempFilepath returns a temp file path. If outputDir is empty, then a temp folder will be created
func getTempFilepath(name string, outputDir string) (string, error) {
	directory := "./"

	if outputDir == "" {
		tempDir, err := ioutil.TempDir("", "go-c8y_")

		if err != nil {
			return "", fmt.Errorf("Could not create temp folder. %s", err)
		}
		directory = tempDir
	}

	return path.Join(directory, name), nil
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
