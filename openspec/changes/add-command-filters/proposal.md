## Why

Some APIs which extensions wrap do not support the server side filtering an extension command needs (issue #582). For example, Analytics Builder models and template models are returned by the same endpoint and only differ by the `c8y_Template` fragment. Today the extension author has to write a script-based command, or ask users to always add `--filter`.

## What Changes

- Extension API commands accept an optional `filter` property: a list of client side filters using the `--filter` syntax.
- The filters are always applied to the command's response, and are combined with the filters given by the user.
- JSON schema (`extensionCommands.json`) and docs are updated, including a caution about pagination.

## Capabilities

### New Capabilities
- `extension-command-filters`: Client side filters of extension API commands.

### Modified Capabilities
<!-- None -->

## Impact

- Code: `internal/integration/models` (`Command.Filter`), `pkg/cmdparser` (`ApplyCommandFilter`, called from `RuntimeCmd.RunE`).
- Docs: API command groups page.
- No change for existing extensions: the property is optional.
