## Context

`set-session` (the shell helper) runs `c8y sessions login`, which selects a session with a provider (`pkg/cmd/sessions/login/login.manual.go`). The provider type comes from `--provider` or `session.provider.type` (`auto`, `external`, `file`, `env`, `stdin`, `interactive`). The `external` type runs `session.provider.command` (default `c8y sessions set --no-banner --output json`) in `FromExternalProvider`:

- environment: `os.Environ()` plus the values of `session.provider.secrets` (from the environment, otherwise from the pin entry prompt `cfg.PromptSecret`)
- arguments: the command, then `--clear`, `--loginType=...`, `--browserCallback=...`, `--sso-scopes=...`, `--sso-audience=...`, `-v`, `--debug` when set, then the positional args of `c8y sessions login`
- stdin and stderr are connected, stdout is captured and parsed as JSON or dotenv (auto-detected unless `--format` is set)

Binary extensions (`add-binary-extensions`) let an extension ship a binary declared in `extension.yaml`, resolved by `Extension.Executable()`, and parse the manifest once per extension (`Extension.Manifest()`). A provider such as `c8y-session-bitwarden` can already be installed as a binary extension and used with `session.provider.command "c8y session-bitwarden list --folder c8y"`. This change makes that a first-class, discoverable integration.

## Goals / Non-Goals

**Goals:**
- An extension declares its providers and the secrets they need, so users don't configure commands or secrets by hand.
- Select a provider by name, from the command line or a setting, with tab completion.
- Run the provider binary directly: one process, no `PATH` lookup of `c8y`.
- Keep the provider's input and output contract identical to `external` providers, so existing providers only need an `extension.yaml` entry.

**Non-Goals:**
- Script based providers (`commands/` files). Providers must be binaries. A script can still be used with the `external` type.
- Changing the provider output format, or providers that store sessions.
- Changing the `set-session` shell helpers.
- Running providers from extensions that are not installed (for example directly from a repository).

## Decisions

### 1. Manifest format

```yaml
binary:
  path: bin/c8y-session-bitwarden
sessionProviders:
  - name: bitwarden                  # [a-z][a-z0-9-]*
    description: Select a session from a Bitwarden folder
    exec: [list]                     # args before the c8y provider flags. Default: []
    secrets: [BW_SESSION]            # environment variables the provider needs
    format: json                     # json, dotenv or auto (default: auto)
```

`exec` uses the same name and meaning as declared binary commands. Providers are part of `extensions.Manifest` and parsed with it, so no extra file reads.

A provider is ignored with a warning when the manifest has no `binary` section. Duplicate names within one extension keep the first entry and log a warning.

*Alternative considered:* a separate `providers.yaml`, or a convention such as a `session` subcommand. A separate file adds another read per extension for little gain, and a fixed subcommand would not let one binary offer several providers (for example one per vault folder) or declare secrets.

### 2. Referencing providers

The provider reference is `<provider>` or `<extension>/<provider>`, where `<extension>` is the extension name without `c8y-`. A short name must be unique among the installed extensions. An ambiguous short name fails with an error that lists the qualified names. Completion always offers the short name when it is unique, and the qualified names otherwise.

*Alternative considered:* always qualified names. They are unambiguous but long to type, and most users have one provider.

### 3. Selecting the provider

- New provider type `extension` (`config.ProviderTypeExtension`) and setting `session.provider.extension` (the provider reference).
- New flag `c8y sessions login --from-extension <ref>`, which works like the other `--from-*` flags: it selects the `extension` type and overrides the settings.
- Without `--from-*` flags, `session.provider.type extension` uses `session.provider.extension`. To keep the type explicit, setting `session.provider.extension` alone does not change the provider type. The docs show both settings:

  ```sh
  c8y settings update session.provider.type extension
  c8y settings update session.provider.extension bitwarden
  ```

- With type `auto`, an `extension` provider is only used when `--from-extension` is given (the existing auto detection is unchanged).

*Alternative considered:* infer the type from `session.provider.extension` being set. That is one command less, but the default type is `external`, so the result would depend on which of two settings wins, which is hard to explain and to debug.

### 4. Running the provider

The existing `FromExternalProvider` is split into building the command and running it (`runProvider(argv, env, format)`), and both provider types use the runner, so stdin, stderr, output capture, format detection, and logging stay identical.

- argv: `Extension.Executable()`, then `exec`, then the same provider flags as today (`--clear`, `--loginType=...`, `-v`, ...), then the positional args of `c8y sessions login`.
- format: `--format` if given, otherwise the declared `format`, otherwise auto-detection.
- environment:
  1. `os.Environ()` without the session variables (`c8ysession.GetSessionEnvKeys()`), so the credentials of the active session are not passed to a program that is selecting a new one
  2. the extension variables `C8Y_EXTENSION_NAME`, `C8Y_EXTENSION_DIR`, `C8Y_BIN`, `C8Y_EXTENSION_CONTEXT_VERSION` (shared helper with the extension runtime context)
  3. secrets: the declared `secrets`, plus `session.provider.secrets` and `--secrets`. Each value is read from the environment, otherwise from the pin entry prompt, as today. Secrets which can't be read are skipped with a debug message, and the provider reports its own error.
- The binary runs directly on all platforms (no `sh`), like other binary commands.

*Alternative considered:* run `c8y <ext> <exec>` through the extension command tree. It reuses more code, but it is the two-process setup this change removes, and the extension command path would add the full runtime context including the active session.

### 5. Discovery

- Completion for `--from-extension` and for `c8y settings update session.provider.extension` lists the providers of all installed extensions (`ExtensionManager().List()`), with the description as completion help.
- `c8y extensions list` output gains a `sessionProviders` field with the provider names.

No new command is added. A `c8y sessions providers list` command can be added later if needed.

### 6. Errors

- Unknown provider: error listing the available provider references.
- Extension binary missing: error naming the extension and the expected path, and suggesting `c8y extensions update <ext>`.
- Provider exits non-zero: same behavior as `external` providers (the login fails with the provider's exit error; its stderr was already shown).

## Risks / Trade-offs

- [Secrets are requested by extension authors] → Secrets are only read from the user's environment or pin entry, never from session files, and only passed to the declaring extension's binary. The provider docs list the secrets it requests, and `--verbose` logs which secret names were passed (never the values).
- [Removing session variables changes behavior compared with `external` providers] → Only for the new `extension` type. Providers don't need the old session, and a provider which needs a value can declare it in `secrets`.
- [Short names can become ambiguous when another provider is installed] → The error lists the qualified names, and qualified names always work, so scripts and settings can use them.
- [Provider declared but binary missing on some platforms] → The binary extension install fails earlier when the release has no asset for the platform, so this only affects broken local installs. The error message points to the expected path.

## Migration Plan

- No migration. Existing provider settings keep working.
- Provider authors add `binary` and `sessionProviders` to `extension.yaml` and publish releases (see binary extensions). `go install` keeps working for users who don't use extensions.
- Rollback: revert the change. Extensions with `sessionProviders` keep working as binary extensions, as unknown manifest fields are ignored.

## Open Questions

- Should `session.provider.extension` imply the `extension` type when `session.provider.type` was never set by the user? (Proposed: no, see decision 3.)
- Should providers also be able to declare default values for login flags, for example `loginType`?
