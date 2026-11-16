## Why

Session providers (for example [c8y-session-bitwarden](https://github.com/reubenmiller/c8y-session-bitwarden)) are external programs that `set-session` and `c8y sessions login` run to select a session. They are installed separately (`go install`), and users configure each one by hand with `session.provider.command` and `session.provider.secrets`. With binary extensions (change `add-binary-extensions`), a provider can be installed with `c8y extensions install`, and `c8y <ext> list` already works as a provider command. But that runs two c8y processes, depends on which `c8y` is first in `PATH`, still needs the user to configure the secrets, and c8y has no way to show which providers are installed.

## What Changes

- Extensions can declare session providers in `extension.yaml` (`sessionProviders`: name, description, exec, secrets, format). Declaring a provider requires a `binary` section.
- New session provider type `extension`, and a `c8y sessions login --from-extension <provider>` flag. c8y runs the extension's binary directly with the declared `exec` arguments, the declared secrets, and the same flags that external providers get today (`--clear`, `--loginType`, `-v`, ...).
- New setting `session.provider.extension` to select the default extension provider. With `session.provider.type extension`, `set-session` uses it without extra arguments.
- Providers are referenced as `<provider>` when the name is unique, or `<extension>/<provider>` otherwise.
- Tab completion for `--from-extension` and `session.provider.extension` lists the installed providers with their descriptions. `c8y extensions list` shows the providers of each extension.
- The provider process gets the extension variables (`C8Y_EXTENSION_NAME`, `C8Y_EXTENSION_DIR`, `C8Y_BIN`, `C8Y_EXTENSION_CONTEXT_VERSION`) but not the credentials of the currently active session, as it is choosing a new session.
- The existing provider types (`external`, `file`, `env`, `stdin`, `interactive`) are unchanged.

## Capabilities

### New Capabilities
- `extension-session-providers`: Declaring session providers in extensions, selecting them by name, running them during `c8y sessions login`, and listing and completing installed providers.

### Modified Capabilities
<!-- None: the binary extension capabilities are introduced by add-binary-extensions, which is not archived yet. This change only adds requirements and does not change them. -->

## Impact

- Code: `pkg/extensions` (manifest types, `SessionProviders` on the manifest), `pkg/cmd/extension` (list output), `pkg/cmd/sessions/login` (new provider type, flag, and execution), `pkg/config` (new setting and provider type), `pkg/cmd/settings/update` (setting and completion).
- Schema: `tools/schema/extension.json` gains `sessionProviders`.
- Docs: binary commands page (session providers section), `c8y sessions login` docs, and the session provider docs.
- Depends on `add-binary-extensions` (PR #743) being merged and archived.
- Security: secrets declared by an extension are only read from the environment or the configured pin entry, and are only passed to that extension's binary. The active session's credentials are not passed to providers.
