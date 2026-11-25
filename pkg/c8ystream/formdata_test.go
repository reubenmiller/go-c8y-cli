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
