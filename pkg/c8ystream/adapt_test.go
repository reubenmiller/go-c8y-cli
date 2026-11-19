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
