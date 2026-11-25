package flags

import (
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	"github.com/spf13/cobra"
)

// TestNewFlagWithPipeIteratorSource checks that a pipeline source replaces
// stdin and the flag values, and that the getter's format is applied to
// non-empty values only.
func TestNewFlagWithPipeIteratorSource(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("device", "flagValue", "")

	source := iterator.NewSliceIterator([]string{"1111", ""})
	iter, err := NewFlagWithPipeIterator(cmd, &PipelineOptions{
		Name:   "device",
		Source: source,
		Format: "id:%s",
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"id:1111", ""} {
		got, _, err := iter.GetNext()
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Errorf("GetNext() = %q, want %q", got, want)
		}
	}

	formatted, err := NewFlagWithPipeIterator(cmd, &PipelineOptions{
		Source:    iterator.NewSliceIterator([]string{"a"}),
		Formatter: func(b []byte) []byte { return append([]byte("x-"), b...) },
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	if got, _, _ := formatted.GetNext(); string(got) != "x-a" {
		t.Errorf("formatter: GetNext() = %q, want x-a", got)
	}
}
