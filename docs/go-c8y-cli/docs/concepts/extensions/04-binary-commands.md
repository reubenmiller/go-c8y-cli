---
category: Concepts - Extensions
title: Binary commands
---

import CodeExample from '@site/src/components/CodeExample';

An extension can provide commands with a compiled binary, for example a program written in Go, Rust, or C#. Binary commands don't need a shell interpreter (so Git for Windows is not required on Windows), they can use any library of their language, and they can be distributed as GitHub releases or with package managers.

A binary extension can also contain API based commands, script based commands, views, templates, and aliases. All commands are merged into the same command tree, so commands from different sources can share a command group.

## Declaring the binary

The binary is declared in the `binary` section of the `extension.yaml` file.

```yaml title="file: extension.yaml"
# yaml-language-server: $schema=https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/v2/tools/schema/extension.json
version: v1

binary:
  # Path to the binary, relative to the extension folder.
  # On Windows, ".exe" is added if the path has no file extension.
  # Default: bin/<extension-name>
  path: bin/c8y-myext

  # Command path under "c8y myext" where the binary is mounted. Default: "" (the extension root)
  mount: ""

  # Forward tab completion to the binary: cobra or none. Default: none
  completion: cobra

# Optional: commands provided by the binary, which are shown in the help
commands:
  - name: services list
    description: List services
    output: json

  - name: ls
    description: Short name for services list
    exec: [services, list]
    output: json
```

### Mount

The binary is mounted at the `mount` path. Any arguments which do not match another command of the extension are passed to the binary unchanged, including flags such as `--help`.

| `mount` | Command | Binary is called with |
|---|---|---|
| `""` | `c8y myext report --month 10` | `report --month 10` |
| `tools` | `c8y myext tools check --verbose` | `check --verbose` |

If the mount path also contains other commands (for example API based commands under `c8y myext devices`), then those commands are used for their paths, and only the other arguments are passed to the binary. In this case `c8y myext --help` shows the c8y help, which lists the other commands. Use `c8y myext help` to show the binary's own help.

### Declared commands

Each entry in `commands` adds a command at `c8y myext <name>`. A name with spaces creates nested command groups, the same way as for API based commands.

| Property | Description |
|---|---|
| `name` | Command path under the extension, for example `services list` |
| `description` | Description shown in the help |
| `exec` | Arguments passed to the binary before the user's arguments. Default: the name split on spaces |
| `output` | `json` if the command writes JSON to stdout (see [JSON output](#json-output)). Default: `raw` |

c8y global flags after a declared command work the same as for built-in commands, for example `c8y myext services list --session prod.json --dry --output csv`. c8y applies them and does not pass them to the binary, so a binary can't use the names of c8y's global flags (for example `--session`, `--output`/`-o`, `--verbose`/`-v`, `--force`/`-f`, `--pageSize`/`-p`) in declared commands. All other flags, `--help`, `-h`, `--examples`, and the arguments after `--` are passed to the binary. Arguments passed to the mount are never changed, so use the mount for commands which need these flag names.

Declared commands appear in the help and can be mixed with API and script commands in the same group. When two sources define the same command, c8y uses the first one in this order: API based command, declared binary command, script based command. Run the command with `--verbose` to see which commands were ignored.

:::note
If `commands` are declared but `extension.yaml` has no `binary` section, then the commands are ignored and a warning is shown. If the binary file does not exist, then a warning is shown and only the other commands of the extension are available.
:::

## JSON output

Commands declared with `output: json` write JSON to stdout, and c8y formats it the same way as its built-in commands. So users can use views, `--select`, `--filter`, and `--output table|csv|json`.

The binary can write:

* a JSON array
* a single JSON object
* newline delimited JSON (one JSON value per line)

The c8y output flags (`--output`/`-o`, `--select`, `--filter`, `--view`, ...) are handled by c8y, like all other [global flags of declared commands](#declared-commands).

If the output is not JSON (for example a message), then it is written unchanged (shown in the log with `--verbose`). The output of `--help` and `-h` is not processed. stderr is never processed, so use it for log and progress messages.

Commands without `output: json`, and arguments passed to the mount, are not processed at all. The binary is connected directly to the terminal, so it can prompt the user.

## Runtime context

c8y passes the following environment variables to all extension commands (binaries and scripts):

| Variable | Description |
|---|---|
| `C8Y_HOST`, `C8Y_TENANT`, `C8Y_USER`, `C8Y_TOKEN`, `C8Y_SESSION`, ... | The active session. These are the same variables that `c8y sessions set` exports |
| `C8Y_HEADER_AUTHORIZATION` | Value of the `Authorization` header for the active session |
| `C8Y_EXTENSION_NAME` | Name of the extension without the `c8y-` prefix |
| `C8Y_EXTENSION_DIR` | Folder where the extension is installed. Use it to read the extension's templates or other files |
| `C8Y_BIN` | Path to the c8y executable which started the command. Use it to call other c8y commands |
| `C8Y_SETTINGS_DEFAULTS_DRY` | `true` when dry run is active (`--dry`) |
| `C8Y_SETTINGS_DEFAULTS_INSECURE` | `true` when TLS verification is turned off (`--insecure`) |
| `C8Y_EXTENSION_NULL_INPUT` | `true` when the extension should not read stdin (`--nullInput` or `-n`). See [Detecting piped input](#detecting-piped-input) |
| `C8Y_SETTINGS_MODE`, `C8Y_MODE` | Session mode, for example `dev` or `prod` (respects `--sessionMode`) |
| `C8Y_EXTENSION_CONTEXT_VERSION` | Version of this list of variables (currently `1`) |

c8y flags placed before the extension name are applied by c8y and are not passed to the extension:

<CodeExample>

```sh
# Run the extension against another session in dry run mode
c8y --session customer1.json --dry myext cleanup
```

</CodeExample>

The password of the session is not passed to extensions. Use the token, the authorization header, or call c8y (`$C8Y_BIN`). c8y does not log in before running an extension, so binary commands also run without an active session. Script based commands still require an active session.

:::caution
Extensions run with your session's credentials, the same as any command you run in your shell after `set-session`. Only install extensions that you trust.
:::

### Detecting piped input

A command which accepts piped input often checks whether stdin is a terminal, for example `[ ! -t 0 ]` in a shell script. That check also matches when there is no terminal and no input either, for example in a cron job, or when the c8y MCP server runs the command. c8y sets `C8Y_EXTENSION_NULL_INPUT=true` when it is run with `--nullInput` (or `-n`), so check it before reading stdin:

<CodeExample>

```sh
if [ ! -t 0 ] && [ "$C8Y_EXTENSION_NULL_INPUT" != "true" ]; then
    # read the piped input
    while IFS= read -r item; do
        echo "$item"
    done
fi
```

</CodeExample>

c8y commands ignore this variable, so pipes between c8y commands within the extension still work. Extension commands which the extension runs inherit it. The c8y MCP server sets `-n` when it runs an extension command without input.

## Tab completion

With `completion: cobra`, c8y forwards tab completion requests to the binary by running it with cobra's hidden `__complete` command. Binaries built with [cobra](https://github.com/spf13/cobra) (including the [Go SDK](#go-sdk)) support this without any extra code. With `completion: none` (the default), the binary is never started for tab completion.

For declared commands, c8y completes the values of its global flags (for example `--output <TAB>`, `--session <TAB>`, and `--view <TAB>`), and forwards the completion of all other arguments to the binary.

## Session providers

An extension can provide [session providers](/docs/concepts/sessions/#activate-a-session-from-a-session-provider), which `set-session` and `c8y sessions login` run to select a session, for example from a password manager.

```yaml title="file: extension.yaml"
binary:
  path: bin/c8y-session-bitwarden

sessionProviders:
  - name: bitwarden                  # lower case letters, numbers and "-"
    description: Select a session from a Bitwarden folder
    exec: [list]                     # arguments before the c8y provider flags
    secrets: [BW_SESSION]            # environment variables the provider needs
    format: json                     # json, dotenv or auto (default)
```

Users select the provider with `c8y sessions login --from-extension bitwarden` (or `set-session --from-extension bitwarden`), or with the `session.provider.type extension` and `session.provider.extension bitwarden` settings.

c8y runs the binary with:

1. the `exec` arguments
2. the login flags which are set: `--clear`, `--loginType=<value>`, `--browserCallback=<value>`, `--sso-scopes=<value>`, `--sso-audience=<value>`, `-v`, `--debug`
3. the arguments the user gave after `--`

For example, `set-session --from-extension bitwarden -v -- --folder c8y` runs `bin/c8y-session-bitwarden list -v --folder c8y`.

The provider must write only the selected session to stdout, in the session file format (json) or as dotenv. Use stderr and stdin to interact with the user, for example to show a list of sessions to pick from. If the provider exits with a non-zero exit code, then the session is not changed.

The provider gets the same environment as the user, except the variables of the active session (`C8Y_HOST`, `C8Y_TOKEN`, `C8Y_SESSION`, ...), plus `C8Y_EXTENSION_NAME`, `C8Y_EXTENSION_DIR`, `C8Y_BIN`, and `C8Y_EXTENSION_CONTEXT_VERSION`. The `secrets` are read from the environment, or from the user's pin entry command if they are not set.

If the provider already verified the credentials (for example it returns a fresh token), then set `"authorized": true` in the session, so c8y doesn't log in again. With the [Go SDK](#go-sdk), use `extsdk.PrintSession`:

```go
authorized := true
return extsdk.PrintSession(extsdk.Session{
    Host:       item.Host,
    Tenant:     item.Tenant,
    Username:   item.Username,
    Password:   item.Password,
    SessionUri: "bitwarden://" + item.ID,
    Authorized: &authorized, // only if the credentials are verified
})
```

## Go SDK

The [extension SDK](https://pkg.go.dev/github.com/reubenmiller/go-c8y-cli/pkg/extsdk) is a small Go module which reads the runtime context, creates a [go-c8y](https://github.com/reubenmiller/go-c8y) client for the active session, and writes JSON output for c8y to format.

```go
ctx, err := extsdk.FromEnv()
client, err := ctx.Client()
devices, _, err := client.Inventory.GetDevices(cmd.Context(), nil)
if err != nil {
    return extsdk.CheckError(err)
}
return extsdk.Print(devices.ManagedObjects)
```

See the [example extension](https://github.com/reubenmiller/go-c8y-cli/tree/v2/pkg/extsdk/examples/c8y-hello), or create a new extension with `c8y extensions create myext --precompiled go`.

## Validating an extension

`c8y extensions validate` checks an extension for problems, for example in the folder you are developing it in:

<CodeExample>

```sh
c8y extensions validate .
```

</CodeExample>

It checks:

| Check | Severity | Problem |
|---|---|---|
| `manifest` | error | `extension.yaml` is not valid YAML, has invalid values (`binary.completion`, `output`), or declares `commands` without a `binary` section, or the same command twice, or session providers which are ignored (no `binary` section, invalid or duplicate name, invalid format) |
| `manifest` | warning | Unknown keys in `extension.yaml`, which usually are typos |
| `name` | error | The declared `name` is invalid (only letters, numbers, and `-`), or is the name of a built-in command |
| `name` | warning | The declared `name` has the `c8y-` prefix or uppercase letters, which are removed or converted |
| `name` | info | An installed extension has another name than it declares, for example because it was installed with `--name` |
| `binary` | error | The declared binary does not exist or is not executable |
| `flags` | warning | A declared command has a flag with the name of a c8y global flag, so c8y handles it and the binary never receives it. This is checked when `binary.completion` is `cobra` |
| `commands` | warning | A command is ignored because another source (API spec, binary, script) defines the same command |
| `api` | error | An API spec can't be parsed |
| `scripts` | warning | A script under `commands/` is not executable |

The exit code is non-zero (101) if there are errors, or warnings when `--strict` is used, so the command can run in CI. The release workflow of `c8y extensions create --precompiled go` runs `c8y extensions validate . --strict` before publishing a release.

`c8y extensions install` also validates the extension (without the `flags` check, which runs the binary) and shows the number of problems.

## Publishing a binary extension

### GitHub releases

`c8y extensions install <owner>/<repo>` checks the latest GitHub release (or the release given by `--pin <tag>`) for an asset for the user's platform. If there is no release, or the release has no assets for any platform, then the repository is installed using `git clone` as before.

Assets are selected in this order:

1. An archive named `<anything>_<os>_<arch>.tar.gz` (or `.zip`), using Go names for the operating system and architecture, for example `c8y-myext_1.0.0_linux_amd64.tar.gz` or `c8y-myext_1.0.0_windows_amd64.zip`. `-` can be used instead of `_`, and `x86_64` and `aarch64` are also accepted.
2. A single binary named `<anything>-<os>-<arch>` (with `.exe` on Windows). This is compatible with GitHub CLI extensions, but can't include views, templates, or other files.
3. On macOS with Apple silicon, the `darwin` `amd64` asset if Rosetta 2 is installed.

An archive must contain `extension.yaml` and the binary at the path declared in it, plus any `views/`, `templates/`, `api/`, or `commands/` folders. The files can be inside a single top-level folder.

```
c8y-myext_1.0.0_linux_amd64.tar.gz
├── extension.yaml
├── bin/
│   └── c8y-myext
└── views/
    └── services.json
```

If the release has a `checksums.txt` file (the [GoReleaser](https://goreleaser.com) format: `<sha256>  <file name>` per line), then c8y verifies the download and stops the install if the checksum does not match. Releases without checksums are installed with a notice.

The extensions created by `c8y extensions create --precompiled go` include a GoReleaser configuration and a GitHub workflow which publish these archives when you push a tag.

`c8y extensions update` installs the latest release, and only replaces the installed version after the download is verified and extracted. Pinned extensions (`--pin`) are not updated. If a git based extension starts publishing release assets, then `c8y extensions update` replaces the git checkout with the release.

To avoid GitHub API rate limits, or to install from private repositories, set `GH_TOKEN` or `GITHUB_TOKEN` (`GH_ENTERPRISE_TOKEN` for GitHub Enterprise Server).

### Package managers

Package managers such as Homebrew, apt, or Scoop install files for all users, so they can't run `c8y extensions install` for each user. Instead, c8y also loads extensions from system-wide folders. A package installs the extension folder (with `extension.yaml`, the binary, and any other folders) into one of them:

| Platform | Folder |
|---|---|
| All | `share/c8y/extensions` next to the c8y executable's folder, for example `/opt/homebrew/share/c8y/extensions` |
| macOS (Homebrew) | `/opt/homebrew/share/c8y/extensions` |
| Linux (Homebrew) | `/home/linuxbrew/.linuxbrew/share/c8y/extensions` |
| Linux | `/usr/local/share/c8y/extensions`, `/usr/share/c8y/extensions` |
| Windows | `%ProgramData%\c8y\extensions` |

For example, a Homebrew formula installs the extension with:

```ruby
def install
  (share/"c8y/extensions/c8y-myext").install "extension.yaml", "views"
  (share/"c8y/extensions/c8y-myext/bin").install "c8y-myext"
end
```

The extension name is the name of the folder without the `c8y-` prefix. If `extension.yaml` declares a `name`, then use `c8y-<name>` as the folder name, so the extension has the same name as when it's installed with `c8y extensions install`.

The folders can be changed with the `extensions.systemDirs` setting (a list separated by `:`, or `;` on Windows):

<CodeExample>

```sh
c8y settings update extensions.systemDirs "/opt/acme/c8y-extensions:/usr/share/c8y/extensions"
```

</CodeExample>

System extensions are shown with the kind `system` in `c8y extensions list`. c8y never changes them: `c8y extensions update` and `c8y extensions delete` show a message to use the package manager instead. If an extension with the same name is installed in your own extension folder, then your copy is used.
