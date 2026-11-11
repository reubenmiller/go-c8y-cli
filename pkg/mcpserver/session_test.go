package mcpserver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionKey(t *testing.T) {
	cases := map[string]string{
		"":                                  "default",
		"/home/me/.cumulocity/dev.json":     "dev",
		"/home/me/.cumulocity/dev.yaml":     "dev",
		"example.c8y.io-first.last":         "example.c8y.io-first.last",
		"/a/example.c8y.io-first.last.json": "example.c8y.io-first.last",
		"/a/example.c8y.io-first.last.env":  "example.c8y.io-first.last",
	}
	for input, want := range cases {
		if got := SessionKey(input); got != want {
			t.Errorf("SessionKey(%q): got=%q, want=%q", input, got, want)
		}
	}
}

func TestResolveSessionFile(t *testing.T) {
	home := t.TempDir()
	file := filepath.Join(home, "example.c8y.io-first.last.json")
	if err := os.WriteFile(file, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(home, "customer", "prod.yaml")
	if err := os.MkdirAll(filepath.Dir(nested), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nested, []byte("host: x"), 0o600); err != nil {
		t.Fatal(err)
	}
	// a file which only starts with the same name should not match
	if err := os.WriteFile(filepath.Join(home, "example.c8y.io-first.last.other.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	cases := map[string]string{
		"example.c8y.io-first.last": file,
		"file://" + file:            file,
		file:                        file,
		"prod":                      nested,
		"unknown":                   "unknown",
		"":                          "",
	}
	for input, want := range cases {
		if got := ResolveSessionFile(input, home); got != want {
			t.Errorf("ResolveSessionFile(%q): got=%q, want=%q", input, got, want)
		}
	}
}
