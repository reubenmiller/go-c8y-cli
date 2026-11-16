---
category: Tutorials - Extensions
title: Creating a binary extension
---

import CodeExample from '@site/src/components/CodeExample';

This tutorial creates an extension with a command written in Go, runs it locally, and publishes it as a GitHub release which users can install with `c8y extensions install`.

For the reference of the `extension.yaml` options, see [binary commands](/docs/concepts/extensions/binary-commands/).

## Prerequisites

* go-c8y-cli
* [Go](https://go.dev/doc/install)
* git
* Optional: [GoReleaser](https://goreleaser.com/install/), to build the release archives locally

## Create the extension

<CodeExample>

```bash
c8y extensions create c8y-inventory-report --precompiled go
cd c8y-inventory-report
```

</CodeExample>

The command creates a git repository with the following files:

```
c8y-inventory-report/
├── .github/workflows/release.yml   # publishes a release when a tag is pushed
├── .goreleaser.yaml                # builds the release archives
├── extension.yaml                  # declares the binary and its commands
├── go.mod
├── main.go                         # example "services list" command
├── README.md
└── views/
    └── services.json               # view for the example items
```

## Run the example command

Build the binary into the `bin/` folder, then install the folder as a local extension. c8y only creates a link to the folder, so you only need to rebuild the binary to see your changes.

<CodeExample>

```bash
go build -o bin/c8y-inventory-report .
c8y extensions install .

c8y inventory-report services list
c8y inventory-report services list --output csv --select name,status
```

</CodeExample>

The `services list` command is declared with `output: json` in `extension.yaml`, so c8y formats its output. The binary only writes JSON:

```go
return extsdk.Print(results)
```

## Add a command which uses the API

Add a command which counts the devices of each type. `extsdk.FromEnv()` reads the session which c8y passes to the extension, and `ctx.Client()` creates a [go-c8y](https://github.com/reubenmiller/go-c8y) client for it.

```go title="file: main.go"
func newTypesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "types",
		Short: "Count the devices of each type",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := extsdk.FromEnv()
			if err != nil {
				return err
			}
			client, err := ctx.Client()
			if err != nil {
				return err
			}
			devices, _, err := client.Inventory.GetDevices(cmd.Context(), &c8y.PaginationOptions{PageSize: 2000})
			if err != nil {
				return extsdk.CheckError(err)
			}

			counts := map[string]int{}
			for _, device := range devices.ManagedObjects {
				counts[device.Type]++
			}
			items := []map[string]any{}
			for name, count := range counts {
				items = append(items, map[string]any{"type": name, "count": count})
			}
			return extsdk.Print(items)
		},
	}
}
```

Register the command in `main()` with `root.AddCommand(newTypesCmd())`, and import `github.com/reubenmiller/go-c8y/pkg/c8y`.

Then declare it in `extension.yaml`, so it is listed in the help and its output is formatted by c8y:

```yaml title="file: extension.yaml"
commands:
  - name: services list
    description: List services
    output: json

  - name: types
    description: Count the devices of each type
    output: json
```

<CodeExample>

```bash
go mod tidy
go build -o bin/c8y-inventory-report .
c8y inventory-report types --output table
```

</CodeExample>

The command uses the active session, so it also works with another session:

<CodeExample>

```bash
c8y --session other-tenant.json inventory-report types
```

</CodeExample>

:::tip
Commands which change data should check `ctx.DryRun()`, so users can preview them with `--dry`. Write messages for the user to stderr, as stdout must only contain JSON for commands with `output: json`.
:::

## Publish a release

Push the repository to GitHub and create a tag. The release workflow builds an archive for each platform with GoReleaser, and publishes them with a `checksums.txt` file.

<CodeExample>

```bash
git add -A
git commit -m "first release"
git tag v1.0.0
git push origin main v1.0.0
```

</CodeExample>

Users can then install the extension. c8y downloads the archive for their platform, verifies its checksum, and installs the binary together with `extension.yaml` and the views.

<CodeExample>

```bash
c8y extensions install <owner>/c8y-inventory-report
```

</CodeExample>

To check the archives before publishing, build them locally:

<CodeExample>

```bash
goreleaser release --snapshot --clean
ls dist/
```

</CodeExample>
