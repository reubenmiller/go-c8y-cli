package extension

import (
	"bytes"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iostreams"
)

func TestCreatePrecompiledCompletion(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cmd := NewCmdExtension(&cmdutil.Factory{IOStreams: ios})

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"__complete", "create", "myext", "--precompiled", ""})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{"go\tGo extension", "other\tExtension in any language", ":4\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("completion output does not contain %q:\n%s", want, got)
		}
	}
}
