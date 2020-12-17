package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/pkg/mapbuilder"
	"github.com/spf13/cobra"
)

const (
	inventoryFlagFragmentType = "fragmentType"
	inventoryFlagQuery        = "query"
	inventoryFlagType         = "type"
	inventoryFlagText         = "text"
	inventoryFlagWithParents  = "withParents"
	inventoryFlagFilter       = "filter"
	inventoryFlagID           = "id"
	inventoryFlagFile         = "file"
)

func addInventoryOptions(cmd *cobra.Command) {
	cmd.Flags().Bool(inventoryFlagWithParents, false, "With parents")
}

func addResultFilterFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(inventoryFlagFilter, "f", "", "Filter property")
}

func addIDFlag(cmd *cobra.Command) {
	cmd.Flags().StringArrayP(inventoryFlagID, "i", []string{}, "Managed Object ID")
	cmd.MarkFlagRequired(inventoryFlagID)
}

func addApplicationFlag(cmd *cobra.Command) {
	cmd.Flags().StringSliceP("application", "i", []string{}, "Application")
	cmd.MarkFlagRequired(inventoryFlagID)
}

func addDataFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(FlagDataName, "d", "", "json")

	// support templating
	addTemplateFlag(cmd)
}

func addDataFlagWithoutTemplates(cmd *cobra.Command) {
	cmd.Flags().StringP(FlagDataName, "d", "", "json")
}

func addTemplateFlag(cmd *cobra.Command) {
	cmd.Flags().String(FlagDataTemplateName, "", "Body template")
	cmd.Flags().String(FlagDataTemplateVariablesName, "", "Body template variables")
}

// resolveTemplatePath resolves a template path
func resolveTemplatePath(name string) (string, error) {
	return matchFilePath(globalFlagTemplatePath, name, ".jsonnet", "ignore")
}

func matchFilePath(sourceDir string, pattern string, extension, ignoreDir string) (string, error) {
	// full path
	if _, err := os.Stat(pattern); err == nil {
		return pattern, nil
	}

	// abort if template path does not exist
	if _, err := os.Stat(sourceDir); err != nil {
		return "", fmt.Errorf("Template path does not exist. %s", sourceDir)
	}

	// path exists under template path
	fullPath := path.Join(sourceDir, pattern)
	if _, err := os.Stat(fullPath); err == nil {
		return fullPath, nil
	}

	// try to resolve path in nested
	names, err := resolvePaths(sourceDir, pattern, extension, ignoreDir)
	if err != nil {
		return "", err
	}

	if len(names) == 0 {
		return "", fmt.Errorf("No matching files found")
	}

	return names[0], nil
}

// resolvePaths find matching files within a directory. The filenames ca be filtered by pattern and extension
func resolvePaths(sourceDir string, pattern string, extension string, ignoreDir string) ([]string, error) {
	files := make([]string, 0)

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() && info.Name() == ignoreDir {
			fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}

		if extension != "" && !strings.HasSuffix(path, extension) {
			return nil
		}

		isMatch := false
		if pattern != "" {
			if matched, _ := filepath.Match(pattern, info.Name()); matched {
				isMatch = true
			}
		}

		if isMatch {
			files = append(files, path)
		}

		return nil
	})
	return files, err
}

func getDataFlag(cmd *cobra.Command) map[string]interface{} {
	if !cmd.Flags().Changed(FlagDataName) {
		return nil
	}
	if value, err := cmd.Flags().GetString(FlagDataName); err == nil {
		return RemoveCumulocityProperties(MustParseJSON(getContents(value)), true)
	}
	return nil
}

func setDataTemplateFromFlags(cmd *cobra.Command, body *mapbuilder.MapBuilder) error {

	if !cmd.Flags().Changed(FlagDataTemplateName) {
		return nil
	}

	if value, err := cmd.Flags().GetString(FlagDataTemplateVariablesName); err == nil {
		content := getContents(value)
		Logger.Infof("Template variables: %s\n", content)
		body.SetTemplateVariables(MustParseJSON(content))
	}

	if value, err := cmd.Flags().GetString(FlagDataTemplateName); err == nil {

		if fullFilePath, err := resolveTemplatePath(value); err == nil {
			Logger.Infof("Template file: %s", fullFilePath)
			value = fullFilePath
		}

		contents := getContents(value)
		body.SetTemplate(contents)
		if err := body.ApplyTemplate(false); err != nil {
			return errors.Wrap(err, "Template error")
		}
	}

	return nil
}

// getContents checks whether the given string is a file reference if so it returns the contents, otherwise it returns the
// input value as is
func getContents(content string) string {
	if _, err := os.Stat(content); err != nil {
		// not a file
		return content
	}

	fileContent, err := ioutil.ReadFile(content)
	if err != nil {
		return content
	}
	// file contents
	return string(fileContent)
}

func getTenantWithDefaultFlag(cmd *cobra.Command, flagName string, defaultTenant string) string {
	if cmd.Flags().Changed(flagName) {
		if value, err := cmd.Flags().GetString(flagName); err == nil {
			return value
		}
	}

	return defaultTenant
}

func getFormDataObjectFlag(cmd *cobra.Command, flagName string, data map[string]interface{}) error {
	if value, err := cmd.Flags().GetString(flagName); err == nil {
		return ParseJSON(value, data)
	}
	return nil
}

func getFileFlag(cmd *cobra.Command, flagName string, formData map[string]io.Reader) error {
	if formData == nil {
		formData = make(map[string]io.Reader)
	}

	// Get custom properties which should be added to the binary
	objectInfo := make(map[string]interface{})
	err := getFormDataObjectFlag(cmd, FlagDataName, objectInfo)
	if err != nil {
		return newSystemErrorF("Could not parse %s flag. %s", FlagDataName, err)
	}

	if filename, err := cmd.Flags().GetString(flagName); err == nil {
		r, err := os.Open(filename)
		if err != nil {
			return errors.New("Failed to read file")
		}

		formData["file"] = r

		if _, ok := objectInfo["name"]; !ok {
			objectInfo["name"] = filepath.Base(filename)
		}

		if _, ok := objectInfo["type"]; !ok {
			mimeType := mime.TypeByExtension(filepath.Ext(filename))

			if mimeType == "" {
				mimeType = "application/octet-stream"
			}

			objectInfo["type"] = mimeType
		}

		if v, err := json.Marshal(objectInfo); err == nil {
			formData["object"] = bytes.NewReader(v)
		} else {
			return errors.New("failed to create object form-data property")
		}

	}
	return nil
}

func getOutputFileFlag(cmd *cobra.Command, flagName string) (filename string, err error) {
	if v, flagErr := cmd.Flags().GetString(flagName); flagErr == nil {
		filename = v
	} else {
		err = newUserError(fmt.Sprintf("Flag [%s] does not exist. %s", flagName, flagErr))
	}
	return
}
