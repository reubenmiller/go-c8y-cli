package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/spf13/cobra"
)

func getFormDataObjectFlag(cmd *cobra.Command, flagName string, data map[string]interface{}) error {
	if value, err := cmd.Flags().GetString(flagName); err == nil {
		return jsonUtilities.ParseJSON(value, data)
	}
	return nil
}

func getFileFlag(cmd *cobra.Command, flagName string, includeMeta bool, formData map[string]io.Reader) error {
	if formData == nil {
		formData = make(map[string]io.Reader)
	}

	// Get custom properties which should be added to the binary
	objectInfo := make(map[string]interface{})
	err := getFormDataObjectFlag(cmd, FlagDataName, objectInfo)
	if err != nil {
		return cmderrors.NewSystemErrorF("Could not parse %s flag. %s", FlagDataName, err)
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

		if includeMeta {
			if v, err := json.Marshal(objectInfo); err == nil {
				formData["object"] = bytes.NewReader(v)
			} else {
				return errors.New("failed to create object form-data property")
			}
		}
	}
	return nil
}
