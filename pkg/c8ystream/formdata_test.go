package c8ystream

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
)

// TestFormParts checks that the file part is named and typed by the form's
// filename and contentType values, which are not sent as parts themselves.
func TestFormParts(t *testing.T) {
	file := filepath.Join(t.TempDir(), "data.bin")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	parts := FormParts(map[string]io.Reader{
		"file":                       f,
		"name":                       strings.NewReader("example"),
		flags.FormDataKeyFilename:    strings.NewReader("custom.bin"),
		flags.FormDataKeyContentType: strings.NewReader("application/x-custom"),
	})
	if len(parts) != 2 {
		t.Fatalf("parts = %v, want file and name only", parts)
	}
	filePart, ok := parts["file"].(*apiv2.FormFile)
	if !ok {
		t.Fatalf("file part = %T, want *apiv2.FormFile", parts["file"])
	}
	if filePart.FileName != "custom.bin" || filePart.ContentType != "application/x-custom" {
		t.Errorf("file part name=%q type=%q, want custom.bin and application/x-custom", filePart.FileName, filePart.ContentType)
	}
	if _, ok := parts["name"].(*apiv2.FormField); !ok {
		t.Errorf("name part = %T, want *apiv2.FormField", parts["name"])
	}

	// Without the metadata values the file name is used
	f2, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()
	parts = FormParts(map[string]io.Reader{"file": f2})
	if got := parts["file"].(*apiv2.FormFile).FileName; got != "data.bin" {
		t.Errorf("file name = %q, want data.bin", got)
	}
}

// TestFormPartsDetectsTypeWithoutParams checks that a file without an explicit
// content type is typed like v2.59.0 did: by the extension of the upload name,
// else by the content, and without parameters ("text/plain", not
// "text/plain; charset=utf-8").
func TestFormPartsDetectsTypeWithoutParams(t *testing.T) {
	dir := t.TempDir()
	open := func(name, content string) *os.File {
		t.Helper()
		file := filepath.Join(dir, name)
		if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close() })
		return f
	}

	tests := []struct {
		name string
		file *os.File
		form map[string]io.Reader
		want string
	}{
		{name: "json extension", file: open("a.json", `{"a":1}`), want: "application/json"},
		{name: "text content without extension", file: open("noext", "hello\n"), want: "text/plain"},
		{name: "binary content without extension", file: open("blob", "\x00\x01\x02\x03"), want: "application/octet-stream"},
		{name: "upload name extension wins", file: open("plain", "hello"), form: map[string]io.Reader{flags.FormDataKeyFilename: strings.NewReader("x.json")}, want: "application/json"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			form := map[string]io.Reader{"file": tc.file}
			for k, v := range tc.form {
				form[k] = v
			}
			got := FormParts(form)["file"].(*apiv2.FormFile).ContentType
			if got != tc.want {
				t.Errorf("content type = %q, want %q", got, tc.want)
			}
		})
	}

	// the sniffed file is still readable from the start
	f := open("again.dat", "hello\n")
	FormParts(map[string]io.Reader{"file": f})
	b, _ := io.ReadAll(f)
	if string(b) != "hello\n" {
		t.Errorf("file contents after detection = %q", b)
	}
}
