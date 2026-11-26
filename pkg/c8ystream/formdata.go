package c8ystream

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
)

// FormParts converts form data built by the flags package (e.g.
// flags.WithFormDataOptions or flags.AddFormFields) to go-c8y multipart parts.
// A file is sent as a file part, named by the flags.FormDataKeyFilename value
// (else by the file name) and typed by the flags.FormDataKeyContentType value
// (else detected from the upload name's extension or the file's content, without
// parameters such as the charset: "text/plain", as v2.59.0 sent). The two metadata values are not sent as parts. Every other
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
			if part.ContentType == "" {
				part.ContentType = detectFileContentType(part.FileName, f)
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

// detectFileContentType returns the media type of an upload without its
// parameters (e.g. "text/plain", not "text/plain; charset=utf-8"): by the
// extension of the upload name, else by sniffing the start of the file (which
// is not consumed). It returns "" when nothing can be determined, so the
// caller's default applies.
func detectFileContentType(fileName string, f *os.File) string {
	if ct := mime.TypeByExtension(filepath.Ext(fileName)); ct != "" {
		return stripMediaParams(ct)
	}
	buf := make([]byte, 512)
	n, _ := f.ReadAt(buf, 0)
	return stripMediaParams(http.DetectContentType(buf[:n]))
}

func stripMediaParams(contentType string) string {
	return strings.TrimSpace(strings.Split(contentType, ";")[0])
}
