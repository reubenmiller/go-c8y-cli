package flags

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFormField(t *testing.T) {
	cases := []struct {
		input    string
		expected FormField
	}{
		{"name=value", FormField{Name: "name", Value: "value"}},
		{"name=", FormField{Name: "name", Value: ""}},
		{"name=a=b", FormField{Name: "name", Value: "a=b"}},
		{"file=@./my.pem", FormField{Name: "file", File: "./my.pem"}},
		{"file=@my.pem;type=text/plain", FormField{Name: "file", File: "my.pem", ContentType: "text/plain"}},
		{"file=@my.pem;filename=other.pem;type=text/plain", FormField{Name: "file", File: "my.pem", Filename: "other.pem", ContentType: "text/plain"}},
		{"file=@my;file.pem", FormField{Name: "file", File: "my;file.pem"}},
		{"text=<notes.txt", FormField{Name: "text", ContentFile: "notes.txt"}},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			field, err := ParseFormField(c.input)
			require.NoError(t, err)
			assert.Equal(t, c.expected, field)
		})
	}
}

func TestParseFormFieldInvalid(t *testing.T) {
	for _, input := range []string{"", "novalue", "=value", "file=@", "file=<"} {
		t.Run(input, func(t *testing.T) {
			_, err := ParseFormField(input)
			assert.ErrorIs(t, err, ErrFlagError)
		})
	}
}

func readAll(t *testing.T, r io.Reader) string {
	t.Helper()
	b, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(b)
}

func TestAddFormFieldsCanBeReopened(t *testing.T) {
	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	notesFile := filepath.Join(dir, "notes.txt")
	require.NoError(t, os.WriteFile(certFile, []byte("cert"), 0o600))
	require.NoError(t, os.WriteFile(notesFile, []byte("notes"), 0o600))

	form := map[string]io.Reader{}
	err := AddFormFields(form,
		"file=@"+certFile+";filename=custom.pem;type=application/x-pem",
		"parentAssetId=1111",
		"notes=<"+notesFile,
	)
	require.NoError(t, err)
	assert.Len(t, form, 5)

	// each request gets fresh readers
	for range 2 {
		values, err := ReopenFormData(form)
		require.NoError(t, err)

		file, ok := values["file"].(*os.File)
		require.True(t, ok, "file fields must be sent as an *os.File")
		assert.Equal(t, "cert", readAll(t, file))
		file.Close()

		_, isFile := values["notes"].(*os.File)
		assert.False(t, isFile, "file contents fields must be sent as text")
		assert.Equal(t, "notes", readAll(t, values["notes"]))
		assert.Equal(t, "1111", readAll(t, values["parentAssetId"]))
		assert.Equal(t, "custom.pem", readAll(t, values[FormDataKeyFilename]))
		assert.Equal(t, "application/x-pem", readAll(t, values[FormDataKeyContentType]))
	}
}

func TestAddFormFieldsErrors(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	require.NoError(t, os.WriteFile(file, []byte("a"), 0o600))

	cases := map[string][]string{
		"duplicate field":         {"a=1", "a=2"},
		"duplicate filename":      {"file=@" + file + ";filename=x", "filename=y"},
		"multiple file options":   {"a=@" + file + ";type=text/plain", "b=@" + file + ";filename=b"},
		"missing file":            {"file=@" + filepath.Join(dir, "missing")},
		"conflict existing field": {"object=1"},
	}
	for name, values := range cases {
		t.Run(name, func(t *testing.T) {
			form := map[string]io.Reader{"object": strings.NewReader("{}")}
			assert.Error(t, AddFormFields(form, values...))
		})
	}
}

func TestMakeFormDataReopenable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	require.NoError(t, os.WriteFile(path, []byte("contents"), 0o600))
	f, err := os.Open(path)
	require.NoError(t, err)

	form := map[string]io.Reader{
		"file":   f,
		"object": strings.NewReader(`{"name":"a.txt"}`),
	}
	require.NoError(t, MakeFormDataReopenable(form))

	for range 2 {
		values, err := ReopenFormData(form)
		require.NoError(t, err)
		file, ok := values["file"].(*os.File)
		require.True(t, ok)
		assert.Equal(t, "contents", readAll(t, file))
		file.Close()
		assert.Equal(t, `{"name":"a.txt"}`, readAll(t, values["object"]))
	}
}
