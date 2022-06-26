package flags

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mapbuilder"
	"github.com/spf13/cobra"
)

var ErrFlagError = errors.New("failed to parse arguments")
var ErrReadFile = errors.New("failed to read file")
var ErrInvalidJSON = errors.New("invalid json")
var ErrUnsupportedType = errors.New("unsupported type")

// WithFormDataOptions returns a body from given command line arguments
func WithFormDataOptions(cmd *cobra.Command, form map[string]io.Reader, inputIterators *RequestInputIterators, opts ...GetOption) (err error) {

	if len(opts) == 0 {
		return nil
	}

	hasInfo := false
	objectInfo := mapbuilder.NewMapBuilder()
	objectInfo.SetEmptyMap()

	for _, opt := range opts {
		name, value, err := opt(cmd, inputIterators)
		if err != nil {
			return err
		}

		if name == "" {
			continue
		}

		switch v := value.(type) {

		case io.Reader:
			if name != "" {
				form[name] = v
			}

		case map[string]interface{}:
			hasInfo = true
			err = objectInfo.MergeMaps(v)

		case Template:
			objectInfo.AppendTemplate(string(v))
			if objectInfo.TemplateIterator == nil {
				objectInfo.TemplateIterator = iterator.NewRangeIterator(1, 100000000, 1)
			}

		case TemplateVariables:
			objectInfo.SetTemplateVariables(v)

		case DefaultTemplateString:
			// the body will build on this template (it can override it)
			objectInfo.PrependTemplate(string(v))

		case RequiredTemplateString:
			// the template will override values in the body
			objectInfo.AppendTemplate(string(v))

		case RequiredKeys:
			objectInfo.SetRequiredKeys(v...)

		default:
			if name != "" {
				err = objectInfo.Set(name, v)
				hasInfo = true

			}
		}
		if err != nil {
			return err
		}
	}

	if hasInfo {
		b, err := objectInfo.MarshalJSON()
		if err != nil {
			return err
		}
		form["object"] = bytes.NewReader(b)
	}

	return nil
}

// WithStringFormValue adds string (as reader) from cli arguments
func WithStringFormValue(opts ...string) GetOption {

	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		src, dst, _ := UnpackGetterOptions("%s", opts...)

		if cmd.Flag(src) == nil {
			return "", nil, nil
		}
		if !cmd.Flags().Changed(src) {
			return "", nil, nil
		}

		value, err := cmd.Flags().GetString(src)

		if err != nil {
			return dst, nil, err
		}

		r := strings.NewReader(value)
		return dst, r, nil
	}
}

// WithFileReader adds file (as reader) from cli arguments
func WithFileReader(opts ...string) GetOption {

	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		src, dst, _ := UnpackGetterOptions("%s", opts...)

		if !cmd.Flags().Changed(src) {
			return "", nil, nil
		}

		value, err := cmd.Flags().GetString(src)

		if err != nil {
			return dst, nil, err
		}

		r, err := os.Open(value)
		return dst, r, err
	}
}

// WithFileBaseName adds the filename basename from cli arguments
func WithFileBaseName(opts ...string) GetOption {

	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		src, dst, _ := UnpackGetterOptions("%s", opts...)

		if !cmd.Flags().Changed(src) {
			return "", nil, nil
		}

		value, err := cmd.Flags().GetString(src)

		if err != nil {
			return dst, nil, err
		}

		return dst, filepath.Base(value), err
	}
}

// WithFileMIMEType adds the file MIME type from cli arguments
func WithFileMIMEType(name string, opts ...string) GetOption {

	return func(cmd *cobra.Command, inputIterators *RequestInputIterators) (string, interface{}, error) {
		src, dst, _ := UnpackGetterOptions("%s", opts...)

		// check for manual value
		if cmd.Flags().Changed(name) {
			typeName, err := cmd.Flags().GetString(name)
			if err == nil {
				return dst, typeName, nil
			}
		}

		if !cmd.Flags().Changed(src) {
			return "", nil, nil
		}

		filename, err := cmd.Flags().GetString(src)

		if err != nil {
			return dst, nil, err
		}

		mimeType := mime.TypeByExtension(filepath.Ext(filename))

		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		return dst, mimeType, err
	}
}

// WithFormDataFile adds form data from cli arguments
func WithFormDataFile(srcFile string, srcData string) []GetOption {
	return []GetOption{
		WithFileReader(srcFile, "file"),
		WithStringFormValue("name", "filename"),
	}
}

// WithFormDataFileAndInfo adds form data from cli arguments
func WithFormDataFileAndInfo(srcFile string, srcData string) []GetOption {
	opts := []GetOption{
		WithFileReader(srcFile, "file"),
		WithFileBaseName(srcFile, "name"),
		WithFileMIMEType("type", srcFile, "type"),
		WithStringFormValue("name", "filename"),
		WithStringValue("name", "name"),
		WithDataValue(srcData),
	}
	return opts
}

func WithFormDataFileAndInfoWithTemplateSupport(templateResolver Resolver, srcFile string, srcData string) []GetOption {
	opts := []GetOption{
		WithFileReader(srcFile, "file"),
		WithFileBaseName(srcFile, "name"),
		WithFileMIMEType("type", srcFile, "type"),
		WithStringFormValue("name", "filename"),
		WithStringValue("name", "name"),
		WithTemplateValue(FlagDataTemplateName, templateResolver),
		WithDataValue(srcData),
	}
	return opts
}
