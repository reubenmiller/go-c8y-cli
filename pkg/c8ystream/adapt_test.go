package c8ystream

import (
	"errors"
	"fmt"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/op"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/spf13/viper"
)

type model struct{ jsondoc.JSONDoc }

func drain(seq output.Seq) (docs []string, errs []error) {
	for d, err := range seq {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		docs = append(docs, string(d.Raw()))
	}
	return
}

func TestAdapt(t *testing.T) {
	r := &Runner{Config: config.NewConfig(viper.New())}
	promptErr := errors.New("User did not confirm action")

	cases := []struct {
		name string
		in   any
		docs int
		errs []error
	}{
		{"nil", nil, 0, nil},
		{"result", op.OK(model{jsondoc.New([]byte(`{"id":"1"}`))}), 1, nil},
		{"no content", op.NoContent(struct{}{}), 0, nil},
		{"error", op.Failed[model](errors.New("boom"), false), 0, []error{errors.New("boom")}},
		{"declined", op.Failed[model](fmt.Errorf("send: %w", apiv2.ErrRequestNotApproved), false), 0, nil},
		{"prompt error", op.Failed[model](fmt.Errorf("Delete \"x\": %w", approvalError{promptErr}), false), 0, []error{promptErr}},
		{"seq", FromDocs(jsondoc.New([]byte(`{}`))), 1, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			docs, errs := drain(r.adapt(c.in))
			if len(docs) != c.docs {
				t.Fatalf("docs = %v, want %d", docs, c.docs)
			}
			if len(errs) != len(c.errs) {
				t.Fatalf("errs = %v, want %v", errs, c.errs)
			}
			for i := range errs {
				if errs[i].Error() != c.errs[i].Error() {
					t.Errorf("err %d = %q, want %q", i, errs[i], c.errs[i])
				}
			}
		})
	}

	_, errs := drain(r.adapt(42))
	if len(errs) != 1 {
		t.Errorf("unsupported type must yield an error, got %v", errs)
	}
}

// A response which is a plain JSON array renders item by item, and as the whole
// body with --raw (as the v2 CLI did). Objects, and a failed result, are not
// affected.
func TestAdaptRawArrayResult(t *testing.T) {
	array := op.OK(model{jsondoc.New([]byte(`[{"a":1},{"a":2}]`))})
	object := op.OK(model{jsondoc.New([]byte(`{"id":"1"}`))})

	for _, c := range []struct {
		name string
		raw  bool
		in   any
		want []string
	}{
		{"array", false, array, []string{`{"a":1}`, `{"a":2}`}},
		{"array raw", true, array, []string{`[{"a":1},{"a":2}]`}},
		{"object raw", true, object, []string{`{"id":"1"}`}},
		{"object", false, object, []string{`{"id":"1"}`}},
	} {
		t.Run(c.name, func(t *testing.T) {
			v := viper.New()
			v.Set(config.SettingsRawOutput, c.raw)
			r := &Runner{Config: config.NewConfig(v)}
			docs, errs := drain(r.adapt(c.in))
			if len(errs) != 0 {
				t.Fatalf("errs = %v", errs)
			}
			if fmt.Sprint(docs) != fmt.Sprint(c.want) {
				t.Errorf("docs = %v, want %v", docs, c.want)
			}
		})
	}
}

// With --raw an enveloped collection renders as its envelope, also when it holds
// no items. An empty top-level array prints "[]" as in the v2 CLI, and --filter
// selects the elements of such an array. --select and --flatten still apply to
// the items of a top-level array, which prints nothing when it is empty.
func TestAdaptRawCollectionResult(t *testing.T) {
	array := op.OK(model{jsondoc.New([]byte(`[{"a":1},{"a":2}]`))})
	emptyArray := op.OK(model{jsondoc.New([]byte(`[]`))})
	envelope := op.WithResponse(op.OK(model{jsondoc.New([]byte(`[{"name":"u1"},{"name":"u2"}]`))}), []byte(`{"users":[{"name":"u1"},{"name":"u2"}]}`))
	// a collection without envelope: the response is the array itself
	arrayResponse := op.WithResponse(op.OK(model{jsondoc.New([]byte(`[{"a":1},{"a":2}]`))}), []byte(`[{"a":1},{"a":2}]`))
	emptyEnvelope := op.WithResponse(op.OK(model{jsondoc.New([]byte(`[]`))}), []byte(`{"users":[]}`))

	for _, c := range []struct {
		name     string
		settings map[string]any
		in       any
		want     []string
	}{
		{"envelope", nil, envelope, []string{`{"name":"u1"}`, `{"name":"u2"}`}},
		{"envelope raw", map[string]any{config.SettingsRawOutput: true}, envelope, []string{`{"users":[{"name":"u1"},{"name":"u2"}]}`}},
		{"envelope raw select", map[string]any{config.SettingsRawOutput: true, config.SettingsSelect: []string{"users.*.name"}}, envelope, []string{`{"users":[{"name":"u1"},{"name":"u2"}]}`}},
		{"empty envelope", nil, emptyEnvelope, nil},
		{"empty envelope raw", map[string]any{config.SettingsRawOutput: true}, emptyEnvelope, []string{`{"users":[]}`}},
		{"empty array", nil, emptyArray, nil},
		{"empty array raw", map[string]any{config.SettingsRawOutput: true}, emptyArray, []string{`[]`}},
		{"empty array raw select", map[string]any{config.SettingsRawOutput: true, config.SettingsSelect: []string{"a"}}, emptyArray, nil},
		{"empty array raw flatten", map[string]any{config.SettingsRawOutput: true, config.SettingsJSONFlatten: true}, emptyArray, nil},
		{"empty array raw filter", map[string]any{config.SettingsRawOutput: true, config.SettingsFilter: []string{"a gt 1"}}, emptyArray, []string{`[]`}},
		{"empty array response raw", map[string]any{config.SettingsRawOutput: true}, op.WithResponse(op.OK(model{jsondoc.New([]byte(`[]`))}), []byte(`[]`)), []string{`[]`}},
		{"array raw select", map[string]any{config.SettingsRawOutput: true, config.SettingsSelect: []string{"a"}}, array, []string{`{"a":1}`, `{"a":2}`}},
		{"array raw filter", map[string]any{config.SettingsRawOutput: true, config.SettingsFilter: []string{"a gt 1"}}, array, []string{`[{"a":2}]`}},
		{"array raw filter no match", map[string]any{config.SettingsRawOutput: true, config.SettingsFilter: []string{"a gt 5"}}, array, []string{`[]`}},
		{"array raw filter select", map[string]any{config.SettingsRawOutput: true, config.SettingsFilter: []string{"a gt 1"}, config.SettingsSelect: []string{"a"}}, array, []string{`{"a":1}`, `{"a":2}`}},
		{"array response raw", map[string]any{config.SettingsRawOutput: true}, arrayResponse, []string{`[{"a":1},{"a":2}]`}},
		{"array response raw select", map[string]any{config.SettingsRawOutput: true, config.SettingsSelect: []string{"a"}}, arrayResponse, []string{`{"a":1}`, `{"a":2}`}},
		{"array raw flatten", map[string]any{config.SettingsRawOutput: true, config.SettingsJSONFlatten: true}, array, []string{`{"a":1}`, `{"a":2}`}},
	} {
		t.Run(c.name, func(t *testing.T) {
			v := viper.New()
			for k, val := range c.settings {
				v.Set(k, val)
			}
			r := &Runner{Config: config.NewConfig(v)}
			docs, errs := drain(r.adapt(c.in))
			if len(errs) != 0 {
				t.Fatalf("errs = %v", errs)
			}
			if fmt.Sprint(docs) != fmt.Sprint(c.want) {
				t.Errorf("docs = %v, want %v", docs, c.want)
			}
		})
	}
}
