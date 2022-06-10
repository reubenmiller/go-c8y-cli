---
title: Aliases
---

import CodeExample from '@site/src/components/CodeExample';

## Overview

go-c8y-cli provides a way to create shortcuts for commands for specific sessions or globally.

Aliases can be created by using the `c8y alias set` command.

Aliases support the following features

* Positional arguments
* Extra arguments
* Optional expanding the alias in shell `sh` (requires `sh` to work)

:::note
Aliases defined by `c8y alias set` are managed by go-c8y-cli and are different to native shell aliases.

Native shell aliases can also be used, however they will only be available in the shell you define them in and are less portable.
:::

## Limitations

Using aliases have the current limitations but maybe lifted in future versions:

* c8y aliases using argument references (i.e. `$1`) are always mandatory and cannot be given a default value

## Examples

### Get a single managed object as json

Create a shortcut to display a single managed object in json

:::info
If you don't use the `--shell` option then you shouldn't prefix the command with `c8y`.
:::

<CodeExample transform="false">

```bash
c8y alias set mo 'inventory get --id "$1" --view off --output json'
```

</CodeExample>

**Usage**

<CodeExample>

```bash
# c8y mo <id>
c8y mo 1234

# Override output format to csv
c8y mo 1234 -o csv
```

</CodeExample>

:::tip
By using `--output json` in the alias, it means the user can override it by specifying another `--output csv`. When parameters are provided twice, then the value form the last parameter wins.
:::

### Lookup by custom name and type

Create a shortcut to lookup a device by a device query

<CodeExample transform="false">

```bash
c8y alias set byName 'devices list --name "$1" --query "has(c8y_IsDevice) and has(devices_IsCustom)" --pageSize 1 --orderBy creationTime.date desc'
```

</CodeExample>

**Usage**

<CodeExample transform="true">

```bash
# c8y byName <device_name>
c8y byName myDevice01

# get device by custom name and chain it with another command
c8y byName myDevice01 | c8y alarms create --template test.alarm.jsonnet
```

</CodeExample>


### Show most recent alarms

Show the most recent 20 alarms which occurred in the last hour

<CodeExample transform="false">

```bash
c8y alias set recentAlarms 'alarms list --dateFrom -1h --pageSize 20'
```

</CodeExample>

**Usage**

<CodeExample>

```bash
c8y recentAlarms
```

</CodeExample>

### Show total number of events for a time period

Show the most recent 20 alarms which occurred in the last hour

<CodeExample transform="false">

```bash
c8y alias set eventCount 'events list --withTotalPages --pageSize 1 --dateFrom "$1" --dateTo "$2"'
```

</CodeExample>

**Usage**

Get the total number of events for a time range

<CodeExample transform="false">

```bash
# c8y eventCount <dateFrom> <dateTo>
c8y eventCount -2h -1h

c8y eventCount "2021-05-01T00:00:00" "2021-05-05T00:00:00" --type "myType"
```

</CodeExample>

### Fail stale executing operations

Fail stale operations created more than x days ago and are still in the EXECUTING state.

The `--shell` option is used here because two commands are going to be chained together:

1. Get the operations
2. Set the piped operations to failed

<CodeExample transform="false">

```bash
c8y alias set --shell failStaleOps 'c8y operations list --dateTo "$1" | c8y operations update --status FAILED --failureReason "User cancelled stale operation"'
```

</CodeExample>

**Usage**

<CodeExample transform="false">

```bash
# c8y failStaleOps <dateTo>
c8y failStaleOps -10d

# Don't prompt for confirmation
c8y failStaleOps -10d --force
```

</CodeExample>


### Fail stale executing operations (extended)

The same fail stale executing operations can be modified to also show the number of operations before the operations are failed.

It involves executing an additional to retrieve the total count using `--pageSize 1 --withTotalPages` trick.

<CodeExample transform="false">

```bash
c8y alias set --shell failStaleOps 'echo -n "Total EXECUTING Operations since $1: "; c8y operations list --dateTo "$1" --withTotalPages --select statistics.totalPages -p 1 -o csv; c8y operations list --dateTo "$1" | c8y operations update --status FAILED --failureReason "User cancelled stale operation"'
```

</CodeExample>

## Settings

The alias are either stored in your session file or global settings file. You can edit them via the command line or by editing the appropriate file.

Below shows how the alias looks in the global session file.


:::note
The aliases which should be expanded on the shell are prefixed with `!`.
:::

```json
{
    "settings": {
        "commonAliases": {
            "recentAlarms": "alarms list --dateFrom -1h",
            "mo": "inventory get --view off --output json --id '$1'"
        },
        "aliases": {
            "searchdevices": "!c8y devices list -p 2000 | grep -i '$1' --color",
            "addTag": "!c8y inventory update --template \"{c8y_tags: if std.objectHas(input.value, 'c8y_tags') then input.value.c8y_tags else {}} + {c8y_tags+:{$1: true}}\""
        }
    }
}
```
