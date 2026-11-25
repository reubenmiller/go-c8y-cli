package cmdparser

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/iterator"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/repository/software/softwareversions"
	"github.com/spf13/cobra"
)

func TestFirstReferenceValue(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringSlice("device", nil, "")
	cmd.Flags().String("name", "", "")
	if err := cmd.Flags().Parse([]string{"--device", "-", "--device", " ,device01,device02", "--name", "-.name"}); err != nil {
		t.Fatal(err)
	}
	if got := firstReferenceValue(cmd, "device", nil); got != "device01" {
		t.Errorf("slice flag: got %q, want device01", got)
	}
	if got := firstReferenceValue(cmd, "name", []string{"fromArg"}); got != "fromArg" {
		t.Errorf("pipeline reference and argument: got %q, want fromArg", got)
	}
	if got := firstReferenceValue(cmd, "missing", nil); got != "" {
		t.Errorf("missing flag: got %q, want empty", got)
	}
}

func TestVersionRef(t *testing.T) {
	ref := softwareversions.NewRef()
	cases := []struct {
		version, parent, want string
	}{
		{version: "12345", parent: "my-app", want: "12345"},
		{version: "1.0.0", parent: "", want: "1.0.0"},
		{version: "1.0.0", parent: "111", want: ref.ByVersion("1.0.0", "111")},
		{version: "1.0.0", parent: "my-app", want: ref.ByVersionAndName("1.0.0", "my-app")},
		{version: "83001,1.0.0", parent: "my-app", want: "83001"},
	}
	for _, tc := range cases {
		got := versionRef(tc.version, tc.parent, ref.ByVersion, func(v, p string) string { return ref.ByVersionAndName(v, p) })
		if got != tc.want {
			t.Errorf("versionRef(%q, %q) = %q, want %q", tc.version, tc.parent, got, tc.want)
		}
	}
}

// TestWithReference checks the reference getter: a flag value is resolved
// once (with the getter's format), the pipeline flag is resolved for each
// item, and an empty piped value is only an error when the flag is required.
func TestWithReference(t *testing.T) {
	refs := &references{client: apiv2.NewClient(apiv2.ClientOptions{BaseURL: "http://127.0.0.1:1"})}
	var lookups []string
	resolve := func(_ context.Context, _ *references, _ *cobra.Command, value string) (string, error) {
		lookups = append(lookups, value)
		if value == "unknown" {
			return "", nil
		}
		return "id-" + value, nil
	}

	cmd := &cobra.Command{}
	cmd.Flags().StringSlice("device", nil, "")
	cmd.Flags().StringSlice("group", nil, "")
	if err := cmd.Flags().Parse([]string{"--group", "My Group,Other"}); err != nil {
		t.Fatal(err)
	}

	source := iterator.NewSliceIterator([]string{"device01", ""})
	inputIterators := &flags.RequestInputIterators{
		PipeOptions: &flags.PipelineOptions{Name: "device", Property: "source.id", Source: source, Format: "ignored %s"},
	}

	// a flag value
	name, value, err := refs.withReference(resolve, nil, "group", "group", "bygroupid(%s)")(cmd, inputIterators)
	if err != nil || name != "group" || value != "bygroupid(id-My Group)" {
		t.Errorf("flag value: got %q=%v, %v", name, value, err)
	}

	// the pipeline flag
	name, value, err = refs.withReference(resolve, nil, "device", "device")(cmd, inputIterators)
	if err != nil || name != "source.id" {
		t.Fatalf("pipeline flag: got %q, %v", name, err)
	}
	iter, ok := value.(iterator.Iterator)
	if !ok || !iter.IsBound() {
		t.Fatalf("pipeline flag: expected a bound iterator. got=%T", value)
	}
	got, _, err := iter.GetNext()
	if err != nil || string(got) != "id-device01" {
		t.Errorf("pipeline item: got %q, %v", got, err)
	}
	got, _, err = iter.GetNext()
	if err != nil || len(got) != 0 {
		t.Errorf("empty pipeline item: got %q, %v", got, err)
	}
	if strings.Join(lookups, ",") != "My Group,device01" {
		t.Errorf("lookups = %v", lookups)
	}

	// a required pipeline flag
	inputIterators.PipeOptions.Required = true
	inputIterators.PipeOptions.Source = iterator.NewSliceIterator([]string{"", "unknown"})
	_, value, _ = refs.withReference(resolve, nil, "device", "device")(cmd, inputIterators)
	iter = value.(iterator.Iterator)
	for i := 0; i < 2; i++ {
		if _, _, err := iter.GetNext(); !errors.Is(err, cmderrors.ErrNoMatchesFound) {
			t.Errorf("item %d: expected no matches found. got=%v", i, err)
		}
	}
}
