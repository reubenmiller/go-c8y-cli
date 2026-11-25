package c8ystream

import (
	"io"
	"os"
	"path/filepath"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
)

// FormParts converts form data built by the flags package (e.g.
// flags.WithFormDataOptions or flags.AddFormFields) to go-c8y multipart parts.
// A file is sent as a file part, named by the flags.FormDataKeyFilename value
// (else by the file name) and typed by the flags.FormDataKeyContentType value
// (else detected). The two metadata values are not sent as parts. Every other
// value is sent as a plain form field.
func FormParts(form map[string]io.Reader) map[string]io.Reader {
	fileName := readFormText(form[flags.FormDataKeyFilename])
	contentType := readFormText(form[flags.FormDataKeyContentType])
	parts := make(map[string]io.Reader, len(form))
	for name, value := range form {
		switch name {
		case flags.FormDataKeyFilename, flags.FormDataKeyContentType:
			continue
		}
		if f, ok := value.(*os.File); ok {
			part := &apiv2.FormFile{Reader: f, FileName: filepath.Base(f.Name()), ContentType: contentType}
			if fileName != "" {
				part.FileName = fileName
			}
			parts[name] = part
			continue
		}
		parts[name] = &apiv2.FormField{Reader: value}
	}
	return parts
}

func readFormText(r io.Reader) string {
	if r == nil {
		return ""
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return ""
	}
	return string(b)
}
