## 0. Prerequisite

- [x] 0.1 Build on `add-binary-extensions` (PR #743): stacked on branch `feat/binary-extensions` while #743 is open. Archive `add-binary-extensions` before archiving this change

## 1. Manifest

- [x] 1.1 Add `SessionProvider` (`name`, `description`, `exec`, `secrets`, `format`) and `Manifest.SessionProviders` to `pkg/extensions`; regenerate mocks if the interface changes
- [x] 1.2 Validate providers when the manifest is used: ignore providers without a `binary` section (warning), keep the first of duplicate names (warning), ignore invalid names (warning)
- [x] 1.3 Add `sessionProviders` to `tools/schema/extension.json`
- [x] 1.4 Unit tests for parsing and validation

## 2. Provider registry

- [x] 2.1 Add a resolver that lists the providers of all installed extensions and resolves `<provider>` and `<extension>/<provider>` references (unique short names, ambiguous and unknown errors that list the available references)
- [x] 2.2 Unit tests with extension mocks: unique, qualified, ambiguous, unknown, extension without binary

## 3. Login integration

- [x] 3.1 Add `config.ProviderTypeExtension`, the `session.provider.extension` setting (env binding, `c8y settings update` entry with completion), and the `extension` value for `session.provider.type`
- [x] 3.2 Add the `--from-extension` flag to `c8y sessions login`, include it in the `--from-*` handling, and select the provider from `session.provider.extension` for type `extension` (error if empty)
- [x] 3.3 Split `FromExternalProvider` into building the command and a shared runner (stdin, stderr, stdout capture, format detection, logging); keep the `external` behavior identical
- [x] 3.4 Implement `FromExtensionProvider`: binary path (error if missing), `exec`, common provider flags, positional args, format precedence (`--format`, declared, auto)
- [x] 3.5 Build the provider environment: remove session variables (`c8ysession.GetSessionEnvKeys()`), add the extension variables (shared helper with the extension runtime context), add secrets from the provider, `session.provider.secrets`, and `--secrets` (environment, then pin entry); log secret names only
- [x] 3.6 Unit tests: argument order, format precedence, environment (session variables removed, declared secret passed, pin entry fallback), missing binary, non-zero exit, `external` provider unchanged

## 4. Discovery

- [x] 4.1 Completion for `--from-extension` and `c8y settings update session.provider.extension` (short names when unique, qualified otherwise, description as help)
- [x] 4.2 Add `sessionProviders` to the `c8y extensions list` output
- [x] 4.3 Tests for completion and list output

## 5. Integration tests and docs

- [x] 5.1 Add a provider to the SDK example extension (or a small test provider) that prints a session from its arguments and environment, without contacting Cumulocity
- [x] 5.2 Commander tests: `--from-extension` with `--format json` and dry handling, settings based selection, completion, unknown and ambiguous references
- [x] 5.3 Document session providers on the binary commands page (manifest, secrets, environment, settings), and update the `c8y sessions login` examples and the session provider docs
- [x] 5.4 Add an `extsdk` helper or example for writing a provider (print the selected session as JSON)
- [x] 5.5 Regenerate the CLI docs for `c8y sessions login` and `c8y settings update`
