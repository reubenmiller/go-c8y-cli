package c8ystream

import (
	"net/http"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/spf13/viper"
)

func testSubmission(semanticMethod string, forceConfirm bool) *submission {
	v := viper.New()
	v.Set(config.SettingsModeConfirmation, "PUT POST DELETE")
	v.Set(config.SettingsForceConfirm, forceConfirm)
	// independent of the environment (e.g. C8Y_SETTINGS_CI of the offline tests)
	v.Set(config.SettingsModeCI, false)
	v.Set(config.SettingsForce, false)
	v.Set(config.SettingsDryRun, false)
	return &submission{cfg: config.NewConfig(v), semanticMethod: semanticMethod}
}

// The semantic method of a command, not the method of the request, decides
// whether a request asks (v2.59.0): a read semantic method only asks with
// --confirm, also when the request is a POST (datahub query,
// deviceregistration getCredentials).
func TestSubmissionRequiresPrompt(t *testing.T) {
	post, _ := http.NewRequest(http.MethodPost, "http://example.com/service/datahub/sql", nil)
	tests := []struct {
		name           string
		semanticMethod string
		forceConfirm   bool
		want           bool
	}{
		{"request method in the confirmation set", "", false, true},
		{"read semantic method", "GET", false, false},
		{"read semantic method with --confirm", "GET", true, true},
		{"write semantic method", "POST", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testSubmission(tt.semanticMethod, tt.forceConfirm).requiresPrompt(post); got != tt.want {
				t.Errorf("requiresPrompt = %v, want %v", got, tt.want)
			}
		})
	}
}

// A request which does not need a confirmation proceeds without a prompt.
func TestSubmissionConfirmWithoutPrompt(t *testing.T) {
	s := testSubmission("GET", false)
	post, _ := http.NewRequest(http.MethodPost, "http://example.com/notification2/token", nil)
	proceed, err := s.confirm(post)
	if !proceed || err != nil {
		t.Errorf("confirm = (%v, %v), want (true, nil)", proceed, err)
	}
	if s.jobID != 0 {
		t.Errorf("a prompt was counted (job %d)", s.jobID)
	}
}
