---
category: Concepts - Extensions - API based commands
title: Command groups
---

A command group is used to logically group commands of a similar nature together.

For example the following in-built commands related to devices is grouped together under a command group called `devices`.

```sh
c8y devices list
c8y devices create
c8y devices update
c8y devices delete
c8y devices get
```

All commands within one API specification are grouped together under the same subcommand. The name of the command group is defined inside the API specification.

:::tip
It is advised to keep the name of the specification file the same as the group name. This makes it easier to maintain the plugin and find the corresponding spec files based on the command group name.
:::

### Example: Multiple commands

The example below defines a command group called `unicorns` and it has three commands. For simplicity all three commands just send a GET request to hardcoded endpoints (including both path and query parameters).

```yaml title="file: api/unicorns.yaml"
# yaml-language-server: $schema=https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/feat/extensions-manager/tools/schema/extensionCommands.json
---
group:
  name: unicorns
  description: Fetch different unicorns

commands:
  # Sub command
  - name: get-command1
    description: Get command 1
    method: GET
    path: inventory/managedObjects?type=pink
  
  # Sub command
  - name: get-command2
    description: Get command 2
    method: GET
    path: inventory/managedObjects?type=yellow
  
  # Sub command
  - name: get-command3
    description: Get command 3
    method: GET
    path: inventory/managedObjects?type=blue
```

The commands are then callable under the extension which they below to, for example assuming it is under an extension called `myextension`, then you can list the available commands using the global  `--help` flag:

<CodeExample>

```sh
c8y myextension unicorns --help
```

</CodeExample>

```bash title="Output"
Example commands

Usage:
  c8y myextension unicorns [command]

Available Commands:
  get-command1 Get command 1
  get-command2 Get command 2
  get-command3 Get command 3

Flags:
  -h, --help   help for example
```
