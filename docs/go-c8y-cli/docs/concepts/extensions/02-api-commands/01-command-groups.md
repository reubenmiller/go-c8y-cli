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
# yaml-language-server: $schema=https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/v2/tools/schema/extensionCommands.json
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

### Example: Ignoring or mapping the pageSize query parameter value

The example below shows how the common query parameter `pageSize` can be controlled to either map the value to a different query parameter, or to ignore the value entirely from the outgoing HTTP request.

By default go-c8y-cli will automatically add the `pageSize` query parameter to all GET requests. This is done as all of the core Cumulocity API supports pagination, so it just made sense to support it by default. Having a default value also allows a global pageSize value to be set in the settings which is then added to all requests automatically.

However, if you are interacting with a service which does not support pagination, then the HTTP request might fail if there are unexpected query parameters sent along with the request.

The example below shows simple commands which execute a GET request. The both are configured to handle the `pageSize` flag value slightly different. The first command will ignore the pageSize value completely, and the seconds command will map the pageSize to a different query parameter. The mapping is made possible by using the `flagMapping` object.

```yaml title="file: api/unicorns.yaml"
# yaml-language-server: $schema=https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/v2/tools/schema/extensionCommands.json
---
group:
  name: unicorns
  description: Fetch different unicorns

commands:
  - name: list
    description: List unicorns
    method: GET
    path: service/unicorns/list
    flagMapping:
      pageSize:       # <== If the value is empty, then the `pageSize` query parameter will be ignored
  
  - name: listpaginated
    description: List unicorns that supports pagination
    method: GET
    path: service/unicorns/paginated
    flagMapping:
      pageSize: limit       # <== Map the `pageSize` value to the `limit` query parameter
```

The first `list` command will exclude the `pageSize` query parameter which is added by default when the `defaults.pageSize` setting is used (or if the user provides `--pageSize <size>`). We can check this behaviour by using dry mode.

<CodeExample>

```sh
c8y settings update defaults.pageSize 100
c8y myextension unicorns list --dry
```

</CodeExample>

As you can see below, the outgoing request does not have the `pageSize` query parameter set.

```bash title="Output"
What If: Sending [GET] request to [https://{host}service/unicorns/list]

### GET service/unicorns/list
```

The second command, `listpaginated`, supports pagination however the unicorn service expects the page size information to be provided via the `limit` query parameter instead of the `pageSize`. The user can control where the `--pageSize` (or default page size) value is mapped to. This enables a more consistent user interface where any differences between services can be normalized on the command line.

Executing the second command using an explicit `--pageSize <size>` flag results in the following output (note `limit=11` is being used now):

<CodeExample>

```sh
c8y myextension unicorns listpaginated --pageSize 11
```

</CodeExample>

```bash title="Output"
What If: Sending [GET] request to [https://{host}/service/unicorns/paginated?limit=11]

### GET /service/unicorns/paginated?limit=11
```
