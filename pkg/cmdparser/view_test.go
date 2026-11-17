package cmdparser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/integration/models"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/spf13/viper"
)

func newViewTestFactory(t *testing.T, viewOption string) (*cmdutil.Factory, *dataview.DataView) {
	t.Helper()
	dir := t.TempDir()
	views := `{
		"version": "1",
		"definitions": [
			{"name": "health", "autoMatch": false, "columns": ["ext.status"]},
			{"name": "global", "autoMatch": false, "columns": ["global.status"]}
		]
	}`
	if err := os.WriteFile(filepath.Join(dir, "views.json"), []byte(views), 0644); err != nil {
		t.Fatal(err)
	}
	otherDir := t.TempDir()
	otherViews := `{"version": "1", "definitions": [{"name": "health", "autoMatch": false, "columns": ["other.status"]}]}`
	if err := os.WriteFile(filepath.Join(otherDir, "views.json"), []byte(otherViews), 0644); err != nil {
		t.Fatal(err)
	}

	// Views of another extension are loaded first, so they would be found first without the namespace
	dv, err := dataview.NewDataView(".*", ".json", nil, "other::"+otherDir, "myext::"+dir)
	if err != nil {
		t.Fatal(err)
	}

	v := viper.New()
	if viewOption != "" {
		v.Set(config.SettingsViewOption, viewOption)
	}
	cfg := config.NewConfig(v)
	log := logger.NewDummyLogger("test")
	return &cmdutil.Factory{
		Config:   func() (*config.Config, error) { return cfg, nil },
		Logger:   func() (*logger.Logger, error) { return log, nil },
		DataView: func() (*dataview.DataView, error) { return dv, nil },
	}, dv
}

func activeColumns(dv *dataview.DataView) []string {
	if view := dv.GetActiveView(); view != nil {
		return view.Columns
	}
	return nil
}

func TestApplyCommandView(t *testing.T) {
	tests := []struct {
		name       string
		viewOption string
		view       *models.CommandView
		want       []string
	}{
		{name: "inline columns", viewOption: "auto", view: &models.CommandView{Columns: []string{"status"}}, want: []string{"status"}},
		{name: "columns take precedence over name", viewOption: "auto", view: &models.CommandView{Name: "health", Columns: []string{"status"}}, want: []string{"status"}},
		{name: "name prefers command's extension", viewOption: "auto", view: &models.CommandView{Name: "health"}, want: []string{"ext.status"}},
		{name: "name with explicit namespace", viewOption: "auto", view: &models.CommandView{Name: "other::health"}, want: []string{"other.status"}},
		{name: "name falls back to all views", viewOption: "auto", view: &models.CommandView{Name: "global"}, want: []string{"global.status"}},
		{name: "unknown name uses auto detection", viewOption: "auto", view: &models.CommandView{Name: "unknown"}, want: nil},
		{name: "no view", viewOption: "auto", view: nil, want: nil},
		{name: "user turned views off", viewOption: "off", view: &models.CommandView{Columns: []string{"status"}}, want: nil},
		{name: "user selected a view", viewOption: "global", view: &models.CommandView{Columns: []string{"status"}}, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, dv := newViewTestFactory(t, tt.viewOption)
			cfg, _ := factory.Config()
			spec := models.Command{Name: "health", Extension: "myext", View: tt.view}
			if err := ApplyCommandView(factory, cfg, spec); err != nil {
				t.Fatal(err)
			}
			got := activeColumns(dv)
			if len(got) != len(tt.want) {
				t.Fatalf("got=%v, want=%v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got=%v, want=%v", got, tt.want)
				}
			}
		})
	}
}

func TestParseSpecWithView(t *testing.T) {
	spec, err := ParseSpec(strings.NewReader(`
group:
  name: analytics
commands:
  - name: health
    method: GET
    path: /service/cep/health
    view:
      columns:
        - status
        - liveness state:components.livenessState
`))
	if err != nil {
		t.Fatal(err)
	}
	view := spec.Commands[0].View
	if view.IsEmpty() || len(view.Columns) != 2 || view.Columns[1] != "liveness state:components.livenessState" {
		t.Errorf("unexpected view. got=%+v", view)
	}
}
