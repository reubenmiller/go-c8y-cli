# extsdk: Go SDK for c8y binary extensions

`extsdk` helps you write [go-c8y-cli](https://goc8ycli.netlify.app) (c8y) extensions as compiled Go binaries.

```sh
go get github.com/reubenmiller/go-c8y-cli/pkg/extsdk
```

The module only depends on [go-c8y](https://github.com/reubenmiller/go-c8y) and [cobra](https://github.com/spf13/cobra), not on the c8y CLI module.

## What it provides

| Function | Description |
|---|---|
| `extsdk.FromEnv()` | Reads the runtime context that c8y passes to extension commands: session, extension folder, path to c8y, dry run, and session mode. Returns `ErrNotRunByC8y` if the binary is not run through `c8y`. |
| `ctx.Client()` | Creates a go-c8y client for the active session. |
| `extsdk.CheckError(err)` | Adds a hint to log in again when a request fails with 401. |
| `extsdk.Print(v)` | Writes JSON to stdout. A slice is written as one JSON value per line. Use it for commands declared with `output: json`, so c8y can format the output (views, `--select`, `--output table`). |
| `extsdk.Exec(ctx, args...)` | Runs a c8y command with the same c8y executable and session. |
| `ctx.ExtensionFile(path)` | Returns the path of a file in the installed extension, for example a template. |
| `ctx.DryRun()` | True when c8y runs in dry run mode (`--dry`). |
| `extsdk.NewRootCommand(name)` | Creates a cobra root command that works with `binary.completion: cobra`. |
| `extsdk.Execute(root)` | Runs the root command and sets the exit code. |

## Example

See [examples/c8y-hello](./examples/c8y-hello). It contains an `extension.yaml` with declared binary commands, an API specification in the same command group, and a view.

```sh
cd examples/c8y-hello
go build -o bin/c8y-hello .
c8y extensions install .
c8y hello services list --output table
```

## Releases

The SDK is a nested Go module, so it has its own tags with the `pkg/extsdk/` prefix, for example `pkg/extsdk/v0.1.0`. The following command runs the SDK tests, creates the tag, and pushes it to `origin` (use `REMOTE=<name>` for another remote):

```sh
task release-extsdk -- v0.1.0
```

Only the git tag is needed. Don't create a GitHub release for it, as the c8y install scripts use the latest GitHub release of the repository.

The runtime context version (`extsdk.ContextVersion`) must match the newest version that c8y sends (`C8Y_EXTENSION_CONTEXT_VERSION`). When c8y adds or changes environment variables, increase the version in both places and release the SDK together with the CLI.
