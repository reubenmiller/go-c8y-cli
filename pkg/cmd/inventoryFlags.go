package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"

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
}

func getDataFlag(cmd *cobra.Command) map[string]interface{} {
	if value, err := cmd.Flags().GetString(FlagDataName); err == nil {
		return RemoveCumulocityProperties(MustParseJSON(getContents(value)), true)
	}
	return make(map[string]interface{})
}

func setDataTemplateFromFlags(cmd *cobra.Command, body *mapbuilder.MapBuilder) error {

	if value, err := cmd.Flags().GetString(FlagDataTemplateVariablesName); err == nil {
		body.SetTemplateVariables(MustParseJSON(getContents(value)))
	}

	if value, err := cmd.Flags().GetString(FlagDataTemplateName); err == nil {
		body.SetTemplate(value)
		body.ApplyTemplate(false)
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

func getFormDataObjectFlag(cmd *cobra.Command, flagName string, formData map[string]io.Reader) error {
	if formData == nil {
		return fmt.Errorf("formData can not be nil")
	}

	if value, err := cmd.Flags().GetString(FlagDataName); err == nil {
		data := MustParseJSON(value)

		if metadataBytes, err := json.Marshal(data); err == nil {
			formData["object"] = bytes.NewReader(metadataBytes)
		}
	}
	return nil
}

func getFileFlag(cmd *cobra.Command, flagName string, formData map[string]io.Reader) error {
	if formData == nil {
		formData = make(map[string]io.Reader)
	}

	if filename, err := cmd.Flags().GetString(flagName); err == nil {
		r, err := os.Open(filename)
		if err == nil {
			formData["file"] = r

			// Add required object field if it does not already exist
			if _, ok := formData["object"]; !ok {
				objectInfo := make(map[string]interface{})
				objectInfo["type"] = mime.TypeByExtension(filepath.Ext(filename))
				objectInfo["name"] = filepath.Base(filename)
				if v, err := json.Marshal(objectInfo); err == nil {
					formData["object"] = bytes.NewReader(v)
				} else {
					return errors.New("failed to create object form-data property")
				}
			}
		} else {
			return errors.New("Failed to read file")
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
