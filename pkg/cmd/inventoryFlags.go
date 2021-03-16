package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/spf13/cobra"
)

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

	_ = cmd.RegisterFlagCompletionFunc(FlagDataTemplateName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		templatePath := cliConfig.GetTemplatePath()
		Logger.Debugf("Template path: %s", templatePath)

		matches, err := resolvePaths(templatePath, "*"+toComplete+"*", ".jsonnet", "ignore")
		for i, match := range matches {
			matches[i] = filepath.Base(match)
		}
		Logger.Debugf("Found: toComplete=%s, total=%d, err=%s", toComplete, len(matches), err)

		if err != nil {
			return []string{"jsonnet"}, cobra.ShellCompDirectiveFilterFileExt
		}
		return matches, cobra.ShellCompDirectiveDefault
	})
}

func addProcessingModeFlag(cmd *cobra.Command) {
	cmd.Flags().String(FlagProcessingModeName, "", "Processing mode")
	completion.WithOptions(
		cmd,
		completion.WithValidateSet(FlagProcessingModeName, "PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP"),
	)
}

type TemplatePathResolver struct {
	Path string
}

func (t *TemplatePathResolver) Resolve(name string) (string, error) {
	return matchFilePath(t.Path, name, ".jsonnet", "ignore")
}

func WithDataValue() flags.GetOption {
	return flags.WithDataValue(FlagDataName, "")
}

func WithTemplateValue() flags.GetOption {
	resolve := &TemplatePathResolver{
		Path: cliConfig.GetTemplatePath(),
	}
	return flags.WithTemplateValue(FlagDataTemplateName, resolve)
}

func WithTemplateVariablesValue() flags.GetOption {
	return flags.WithTemplateVariablesValue(FlagDataTemplateVariablesName)
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
