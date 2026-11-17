package cmdparser

import (
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/internal/integration/models"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/dataview"
)

// ApplyCommandView activates the default view of an extension API command.
// The view is only applied when automatic view detection is active, so
// it does not override views selected by the user (e.g. --view <name> or --view off).
// The --select flag is not affected, as it already takes precedence over views
func ApplyCommandView(factory *cmdutil.Factory, cfg *config.Config, spec models.Command) error {
	if spec.View.IsEmpty() || factory.DataView == nil {
		return nil
	}

	if !strings.EqualFold(cfg.ViewOption(), config.ViewsAuto) {
		return nil
	}

	dataView, err := factory.DataView()
	if err != nil || dataView == nil {
		return err
	}

	log, err := factory.Logger()
	if err != nil {
		return err
	}

	definition, err := ResolveCommandView(dataView, spec)
	if err != nil {
		return err
	}
	if definition == nil {
		log.Warnf("Command view not found, so the view will be detected automatically. command=%s, view=%s", spec.Name, spec.View.Name)
		return nil
	}

	log.Debugf("Using the command's view. command=%s, view=%s", spec.Name, definition.FQDN())
	dataView.SetActiveView(definition)
	return nil
}

// ResolveCommandView returns the view definition of an extension API command.
// Columns take precedence over the view name. A view name is looked up in
// the command's extension first, and then in all views.
// nil is returned if the view name does not match any view
func ResolveCommandView(dataView *dataview.DataView, spec models.Command) (*dataview.Definition, error) {
	if spec.View.IsEmpty() {
		return nil, nil
	}

	if len(spec.View.Columns) > 0 {
		return &dataview.Definition{
			Name:      spec.Name,
			Extension: spec.Extension,
			Columns:   spec.View.Columns,
		}, nil
	}

	if spec.Extension != "" && !strings.Contains(spec.View.Name, dataview.NamespaceSeparator) {
		definition, err := dataView.GetDefinitionByName(spec.Extension + dataview.NamespaceSeparator + spec.View.Name)
		if err != nil || definition != nil {
			return definition, err
		}
	}
	return dataView.GetDefinitionByName(spec.View.Name)
}
