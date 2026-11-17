## Why

An extension's name, and so its command `c8y <name>`, comes from the repository name without the `c8y-` prefix. Authors have to name repositories like `c8y-dtm` or `c8y-analytics`, which are too generic to find or to tell apart on GitHub (issue #579). Users can already pass `c8y extensions install --name`, but then each user has to know and type the intended name.

## What Changes

- `extension.yaml` gets an optional top-level `name`, the extension name without the `c8y-` prefix. For example, the repository `c8y-cli-extension-analytics` can declare `name: analytics` and is installed as `c8y-analytics`, with the command `c8y analytics`.
- `c8y extensions install` resolves the name in this order: `--name`, then `name` in `extension.yaml`, then the repository (or folder) name as today. This applies to git installs, binary installs from GitHub releases, and local installs (`c8y extensions install .`).
- c8y downloads (or clones) the extension into a temporary folder first, reads the name from its `extension.yaml`, and checks the name against built-in commands and installed extensions before moving it into place. Git installs now use a temporary folder too, as binary installs already do.
- A repository whose `extension.yaml` declares a name doesn't need the `c8y-` prefix in the repository name.
- An installed extension keeps its name on `c8y extensions update`. If a new version declares a different name, c8y prints a notice that the extension must be reinstalled to use it.
- `c8y extensions create` writes `name` into the generated `extension.yaml` (script, `--precompiled go`, and `--precompiled other` templates).
- `c8y extensions validate` checks the name: invalid format, the name of a built-in command, and (for installed extensions) a name which differs from the installed name.
- The JSON schema `tools/schema/extension.json` gets the `name` property.
- No breaking change: extensions without `name` are named as before, and already installed extensions keep their names.

Not included, as alternatives from the issue discussion: removing a longer prefix such as `c8y-cli-` from repository names, or using the part after the last `-`. Both would change the names of extensions which are already installed or published, and the second gives wrong names for names like `c8y-device-simulator`.

## Capabilities

### New Capabilities

- `extension-manifest-name`: The `name` field in `extension.yaml`, its format, how the installed name is resolved (`--name`, manifest, repository or folder name) for git, binary, and local installs, the repository prefix rule, and the behavior on update.

### Modified Capabilities

- `extension-validate`: New name checks (format, built-in command clash, installed name differs from the declared name).
- `extension-scaffolding`: The generated `extension.yaml` declares the extension name.

## Impact

- `pkg/extensions/extension.go`: `Manifest.Name`.
- `pkg/cmd/extension/command.go`: the `install` command (name resolution, repository prefix check after download), `checkValidExtension`.
- `pkg/cmd/extension/manager.go`: `Install`, `installGit` (clone into a temporary folder), `installBinRelease`, `InstallLocal`, upgrades (name change notice).
- `pkg/cmd/extension/validate.go`: name checks.
- `pkg/cmd/extension/ext_tmpls/`: `extension.yaml` templates.
- `tools/schema/extension.json`: `name` property.
- Docs: extension authoring pages (naming), `c8y extensions install` help.
- Tests: unit tests for name resolution and validation, commander tests for install and validate.
