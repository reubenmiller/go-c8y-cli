package flags

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileguard"
	"github.com/spf13/cobra"
)

// Form data keys which the multipart encoder uses as metadata for the uploaded file
// rather than as standalone form fields
const (
	FormDataKeyFilename    = "filename"
	FormDataKeyContentType = "contentType"
)

// ReopenableReader is a form data value which can create a fresh reader for each request,
// so the same form data can be sent multiple times (e.g. when piping input to a command).
// Read can be used directly, however it can only be consumed once.
type ReopenableReader interface {
	io.Reader
	Reopen() (io.Reader, error)
}

// ReopenFormData returns a copy of the form data where each ReopenableReader is replaced
// by a fresh reader. Other values are copied as is
func ReopenFormData(form map[string]io.Reader) (map[string]io.Reader, error) {
	if len(form) == 0 {
		return form, nil
	}
	out := make(map[string]io.Reader, len(form))
	for key, value := range form {
		if v, ok := value.(ReopenableReader); ok {
			r, err := v.Reopen()
			if err != nil {
				return nil, err
			}
			out[key] = r
		} else {
			out[key] = value
		}
	}
	return out, nil
}

// FormText is a text form field
type FormText struct {
	value []byte
	r     io.Reader
}

// NewFormText creates a new text form field
func NewFormText(value []byte) *FormText {
	return &FormText{value: value}
}

func (f *FormText) Read(p []byte) (int, error) {
	if f.r == nil {
		f.r = bytes.NewReader(f.value)
	}
	return f.r.Read(p)
}

// Reopen returns a new reader for the text value
func (f *FormText) Reopen() (io.Reader, error) {
	return bytes.NewReader(f.value), nil
}

// FormFile is a form field which is uploaded as a file. The file is opened for each request
type FormFile struct {
	Path string
	r    io.Reader
}

// NewFormFile creates a new file form field
func NewFormFile(path string) *FormFile {
	return &FormFile{Path: path}
}

func (f *FormFile) Read(p []byte) (int, error) {
	if f.r == nil {
		file, err := fileguard.Open(f.Path)
		if err != nil {
			return 0, err
		}
		f.r = file
	}
	return f.r.Read(p)
}

// Reopen opens the file. An *os.File is returned so the multipart encoder sends it as a file
func (f *FormFile) Reopen() (io.Reader, error) {
	file, err := fileguard.Open(f.Path)
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrReadFile, err)
	}
	return file, nil
}

// FormFileContents is a text form field whose value is read from a file for each request
type FormFileContents struct {
	Path string
	r    io.Reader
}

// NewFormFileContents creates a text form field which uses the contents of a file
func NewFormFileContents(path string) *FormFileContents {
	return &FormFileContents{Path: path}
}

func (f *FormFileContents) Read(p []byte) (int, error) {
	if f.r == nil {
		r, err := f.Reopen()
		if err != nil {
			return 0, err
		}
		f.r = r
	}
	return f.r.Read(p)
}

// Reopen reads the file contents. A non-file reader is returned so the value is sent as a text field
func (f *FormFileContents) Reopen() (io.Reader, error) {
	b, err := fileguard.ReadFile(f.Path)
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrReadFile, err)
	}
	return bytes.NewReader(b), nil
}

// FormField is a parsed curl style form field, e.g. name=value, name=@file or name=<file
type FormField struct {
	Name        string
	Value       string
	File        string
	ContentFile string
	Filename    string
	ContentType string
}

// Reader returns the form data value
func (f FormField) Reader() io.Reader {
	switch {
	case f.File != "":
		return NewFormFile(f.File)
	case f.ContentFile != "":
		return NewFormFileContents(f.ContentFile)
	default:
		return NewFormText([]byte(f.Value))
	}
}

// ParseFormField parses a curl style form field:
//
//	name=value                                 text field
//	name=@path[;type=mime/type][;filename=name] file upload
//	name=<path                                 text field using the contents of a file
func ParseFormField(v string) (FormField, error) {
	field := FormField{}
	name, value, found := strings.Cut(v, "=")
	name = strings.TrimSpace(name)
	if !found || name == "" {
		return field, fmt.Errorf("%w. invalid form field %q, expected name=value, name=@file or name=<file", ErrFlagError, v)
	}
	field.Name = name

	switch {
	case strings.HasPrefix(value, "@"):
		parts := strings.Split(value[1:], ";")
		pathParts := []string{parts[0]}
		for _, part := range parts[1:] {
			if v, ok := strings.CutPrefix(part, "type="); ok {
				field.ContentType = v
			} else if v, ok := strings.CutPrefix(part, "filename="); ok {
				field.Filename = v
			} else {
				// not a known option, so assume it is part of the path
				pathParts = append(pathParts, part)
			}
		}
		field.File = strings.Join(pathParts, ";")
		if field.File == "" {
			return field, fmt.Errorf("%w. form field %q is missing a file path", ErrFlagError, v)
		}
	case strings.HasPrefix(value, "<"):
		field.ContentFile = value[1:]
		if field.ContentFile == "" {
			return field, fmt.Errorf("%w. form field %q is missing a file path", ErrFlagError, v)
		}
	default:
		field.Value = value
	}
	return field, nil
}

// AddFormFields adds curl style form fields to the form data.
// Each field name can only be used once. The file options (type and filename) can only be used
// by a single file field as the multipart encoder uses them for all files
func AddFormFields(form map[string]io.Reader, values ...string) error {
	fileOptionsUsed := false
	for _, v := range values {
		field, err := ParseFormField(v)
		if err != nil {
			return err
		}

		newKeys := []string{field.Name}
		if field.Filename != "" {
			newKeys = append(newKeys, FormDataKeyFilename)
		}
		if field.ContentType != "" {
			newKeys = append(newKeys, FormDataKeyContentType)
		}
		if field.Filename != "" || field.ContentType != "" {
			if fileOptionsUsed {
				return fmt.Errorf("%w. the form file options 'type' and 'filename' can only be used on one file", ErrFlagError)
			}
			fileOptionsUsed = true
		}
		for _, key := range newKeys {
			if _, exists := form[key]; exists {
				return fmt.Errorf("%w. form field %q is set more than once", ErrFlagError, key)
			}
		}

		if field.File != "" {
			// fail early if the file does not exist
			if _, err := os.Stat(field.File); err != nil {
				return fmt.Errorf("%w. %w", ErrReadFile, err)
			}
			if err := fileguard.Check(field.File); err != nil {
				return err
			}
		}

		form[field.Name] = field.Reader()
		if field.Filename != "" {
			form[FormDataKeyFilename] = NewFormText([]byte(field.Filename))
		}
		if field.ContentType != "" {
			form[FormDataKeyContentType] = NewFormText([]byte(field.ContentType))
		}
	}
	return nil
}

// WithFormFields adds curl style form fields from a string array flag
func WithFormFields(cmd *cobra.Command, form map[string]io.Reader, flagName string) error {
	if cmd.Flags().Lookup(flagName) == nil || !cmd.Flags().Changed(flagName) {
		return nil
	}
	values, err := cmd.Flags().GetStringArray(flagName)
	if err != nil {
		return err
	}
	return AddFormFields(form, values...)
}

// MakeFormDataReopenable converts the form data values so that they can be reopened for each request.
// Files are reopened by their path, and other readers are buffered in memory
func MakeFormDataReopenable(form map[string]io.Reader) error {
	for key, value := range form {
		switch v := value.(type) {
		case ReopenableReader:
			continue
		case *os.File:
			form[key] = NewFormFile(v.Name())
			_ = v.Close()
		default:
			b, err := io.ReadAll(v)
			if err != nil {
				return err
			}
			if c, ok := v.(io.Closer); ok {
				_ = c.Close()
			}
			form[key] = NewFormText(b)
		}
	}
	return nil
}
