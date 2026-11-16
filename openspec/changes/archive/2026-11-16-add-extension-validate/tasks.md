## 1. Validator

- [x] 1.1 Add the finding model (severity, check, location, message) and a validator for an `Extension` with the c8y global flag set as input
- [x] 1.2 Manifest checks: YAML errors, unknown keys with `KnownFields(true)`, value checks, command names, commands without binary
- [x] 1.3 Binary checks: exists, regular file, executable (not on Windows)
- [x] 1.4 Flag clash checks with `__complete` (timeout, extension variables), info when completion is not cobra, warning when reading fails
- [x] 1.5 Command tree checks: build a `cmdtree.Tree` with API, binary, mount, and script paths and collect conflicts
- [x] 1.6 API spec parse errors and non-executable scripts
- [x] 1.7 Unit tests for each check, including the kitchensink fixture and the SDK example
- [x] 1.8 Session provider checks (from `add-extension-session-providers`): providers which are ignored are errors

## 2. Command

- [x] 2.1 Add `c8y extensions validate [<name> | <path>]` with `--strict`, output through the c8y output handling, success message on stderr, and the exit code
- [x] 2.2 Validate after `c8y extensions install` (without the flag check) and print a summary on stderr when there are problems
- [x] 2.3 Commander tests: clean extension, extension with problems (table and json output, exit codes with and without `--strict`), unknown extension

## 3. Template and docs

- [x] 3.1 Add a validation step to the `--precompiled go` release workflow template
- [x] 3.2 Document validation on the binary commands page, and regenerate the CLI docs for `c8y extensions`
