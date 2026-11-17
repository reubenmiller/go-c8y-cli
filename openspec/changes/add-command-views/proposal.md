## Why

Views are matched globally on the response and request (issue #580). An extension which ships a view for its own endpoint, for example one that matches `requestPath: /service/cep/health`, also changes the output of `c8y api /service/cep/health` and of any other command which calls that path. Extension authors need a way to set the view of their own commands without affecting other commands.

## What Changes

- Extension API commands accept an optional `view` property with either `columns` (an inline view) or `name` (an existing view). The view is used when the view option is `auto`, so `--view <name>`, `--view off`, the `defaults.view` setting, and `--select` still take precedence.
- A view `name` is resolved in the command's extension first, then in all views. `<extension>::<name>` picks a view from a specific extension. An unknown name logs a warning and falls back to automatic detection.
- View definitions accept `"autoMatch": false`, which excludes them from automatic detection, so they are only used by name.
- JSON schemas (`extensionCommands.json`, `views.json`) and docs are updated.

## Capabilities

### New Capabilities
- `extension-command-views`: Default views of extension API commands, and views which are excluded from automatic detection.

### Modified Capabilities
<!-- None -->

## Impact

- Code: `internal/integration/models` (`Command.View`, `Command.Extension`), `pkg/dataview` (`autoMatch`, `SetActiveView`, `GetDefinitionByName`), `pkg/cmdparser` (`ApplyCommandView`), `pkg/cmd/root` (sets the extension name of API commands).
- Docs: API command groups page, views concept page.
- No change for existing extensions or views: both properties are optional and default to the current behavior.
