---
layout: default
category: Configuration
title: Settings
---

import CodeExample from '@site/src/components/CodeExample';

PSc8y and c8y supports a modern approach to configuration. It allows you to control settings via configuration files and/or environment variables.

## Order of processing

The settings are read and applied in the following order:

1. Read `settings.<json|yaml>` inside current working directory
2. Read `settings.<json|yaml>` file inside the directory specified by the `C8Y_SESSION_HOME` env variable
3. Read session file path in the `C8Y_SESSION` env variable
4. Read setting from env variable (if defined)
5. Read from command argument equivalent (if it exists)

The value last set will be used by `c8y`.

You can inspect the current settings interpreted by go-c8y-cli by running:

<CodeExample>

```bash
c8y settings list --select "**" --output json
```

</CodeExample>

```json title="Output"
{
  "activitylog": {
    "currentPath": "/home/exampleuser/demo/activityLog/c8y.activitylog.2021-05-14.json",
    "enabled": true,
    "methodfilter": "GET PUT POST DELETE",
    "path": "$C8Y_HOME/activityLog"
  },
  "aliases": {
    "perf": "!c8y activitylog list --dateFrom '$1' --select responseTimeMS -o csv | datamash -H -W mean 1 max 1 min 1 range 1"
  },
  "ci": false,
  "commonaliases": {
    "mo": "inventory get --view off --output json --id '$1'",
    "recentalarms": "alarms list --dateFrom -1h"
  },
  "defaults": {
    "abortonerrors": 10,
    "cache": false,
    "cachettl": "60s",
    "compact": false,
    "confirm": false,
    "confirmtext": "",
    "currentpage": 0,
    "debug": false,
    "delay": "10ms",
    "delaybefore": "0ms",
    "dry": false,
    "dryformat": "markdown",
    "examples": false,
    "flatten": false,
    "force": false,
    "includeall": false,
    "logmessage": "",
    "maxjobs": 0,
    "maxworkers": 50,
    "noaccept": false,
    "nocolor": false,
    "nolog": false,
    "noproxy": false,
    "nullinput": false,
    "output": "json",
    "outputfile": "",
    "outputfileraw": "",
    "pagesize": 20,
    "progress": false,
    "proxy": "",
    "raw": false,
    "session": "",
    "sessionpassword": "",
    "sessionusername": "",
    "silentexit": false,
    "silentstatuscodes": "",
    "timeout": "60s",
    "totalpages": 0,
    "verbose": false,
    "view": "auto",
    "witherror": false,
    "withtotalpages": false,
    "workers": 1
  },
  "encryption": {
    "cachepassphrase": true,
    "enabled": true
  },
  "includeall": {
    "delayms": 50,
    "pagesize": 2000
  },
  "logger": {
    "hidesensitive": "true"
  },
  "mode": {
    "confirmation": "PUT POST DELETE",
    "enablecreate": true,
    "enabledelete": true,
    "enableupdate": true
  },
  "path": "",
  "session": {
    "defaultusername": "",
  },
  "storage": {
    "storepassword": true,
    "storetoken": true
  },
  "template": {
    "path": "/home/exampleuser/demo/templates"
  },
  "views": {
    "columnmaxwidth": 80,
    "columnminwidth": 5,
    "columnpadding": 5,
    "commonpaths": ["/home/exampleuser/.go-c8y-cli/views/default"],
    "custompaths": ["$C8Y_HOME/views"],
    "rowmode": "truncate"
  }
}
```

:::note
The settings can be played under the `.settings` property of either your global settings file or a current session file.

```jsonc
{
  "settings": {
    // all settings
  }
}
```
:::


## Global settings

### Example: Set global defaults to use in each c8y session

Global settings can be controlled by creating a `settings.json` file in the `$C8Y_HOME` folder, and adding the following contents:

```json title="file: ~/.cumulocity/settings.json"
{
  "$schema": "https://raw.githubusercontent.com/exampleuser/go-c8y-cli/v2/tools/schema/session.schema.json",
  "settings": {
      "storage": {
        "storepassword": false,
        "storetoken": false
      },
      "defaults": {
        "pageSize": 30
      }
  }
}
```

The same settings can also be added to your session file, so that you can override the defaults defined in the `settings.json` file.

Below shows an example of a session (json) file with a `settings` section, where the default pageSize is set to 50.

```json title="file: ~/.cumulocity/my-session01.json"
{
  "$schema": "https://raw.githubusercontent.com/exampleuser/go-c8y-cli/v2/tools/schema/session.schema.json",
  "host": "https://example.zz-latest.cumulocity.com",
  "tenant": "t12345",
  "username": "hans@example.com",
  "password": "h4n$gRu8er",
  "description": "Nakatomi building management",
  "settings": {
      "defaults:" {
        "pageSize": 50
      }
  }
}
```

:::tip
* If the same settings exists in the `settings.json` and session file, then the value in the session file will be used
* The `$schema` property is provided to enable tab completion of the properties when editing in VS Code (or another editor which supports the json schema lookups)
:::

## Special Environment variables

Some settings cannot be defined in the session or settings files as they control the behaviour related to such files.

The following is a list of available environment variables which control how c8y and PSc8y interacts with activating, searching and displaying c8y sessions.

### C8Y_HOME

Directory which will be available for use in other configuration options that expect a path. It allows your folders to be grouped together

### C8Y_SESSION_HOME

By default the `$HOME/.cumulocity` directory is used to store the Cumulocity session files. A custom session home folder can be specified by setting the `C8Y_SESSION_HOME` to a folder.

Use a custom folder where the Cumulocity Session files should be kept and searched through.

## Environment variable mapping

All of the settings can be also set via an environment settings.

The equivalent environment variable name can be calculated by applying the following transformation:

1. Convert values to upper case
2. Replace `.` characters with `_`
3. Add `C8Y_SETTINGS_` to the name

The following shows a few examples of the transformation:

| Setting | Environment Variable |
|------|----------------------|
| `ci` | `C8Y_SETTINGS_CI` |
| `defaults.pageSize` | `C8Y_SETTINGS_DEFAULTS_PAGESIZE` |
| `defaults.force` | `C8Y_SETTINGS_DEFAULTS_FORCE` |

---

## Tab completion 

go-c8y-cli provides a way to edit the configuration from the command line either by setting a value in your configuration file, or setting an environment variable.

<CodeExample>

```bash
c8y settings update <tab><tab>
c8y settings update logger.hideSensitive <tab><tab>
```

</CodeExample>

---

## Settings

### activitylog.enabled: boolean

Enable activity log

<CodeExample>

```bash
c8y settings update activitylog.enabled true
```

</CodeExample>

### activitylog.methodfilter: string

<CodeExample>

```bash
c8y settings update activitylog.methodfilter "GET PUT POST DELETE"
```

</CodeExample>

Types of HTTP Methods to include in the activity log.

You can reduce choose to ignore GET requests from the activity log by using:

<CodeExample>

```bash
c8y settings update activitylog.methodfilter "PUT POST DELETE"
```

</CodeExample>

### activitylog.path: string

Base folder where the activity log files should be stored.

<CodeExample>

```bash
c8y settings update activitylog.path '$C8Y_HOME/activityLog'
```

</CodeExample>

:::tip
`$C8Y_HOME` is a special go-c8y-cli variable that can be used to reference your c8y home folder, so you will need to make sure you use single quotes when setting it via the command line so it does not get expanded to a shell variable.
:::

### defaults.cache: boolean

Enable caching. Defaults to `false`.

Caching can be turned on for individual commands by using the `--cache` global parameter.

When the setting is activated, then all requests which match the `cache.methods` settings will be cached. Users can opt out of caching for individual commands by using the `--noCache` global parameter.

### defaults.cacheTTL: string

Cache time-to-live (TTL) settings to control the maximum age that a cached response. If the cached response is older than the TTL settings, then the cached item will be ignored.

### cache.keyauth: boolean

Include the request's Authorization header value when generating the cache key. Defaults to `true`.

This can be useful if you want to share caching across multiple users. Normally the authorization header is used in the cache key generation, so the cached item will be a different file key for multiple users

### cache.keyhost: boolean

Include the host name when generating the cache key. Defaults to `true`.

This can be useful when setting up a mock system where you want to fake server response by using cached response regardless of the server. This is a more advanced settings, so just ignore it if you don't understand it.

### cache.methods: string

Space separated list of HTTP methods which should be cached. By default only `GET` is configured.

<CodeExample>

```bash
# Enable more http methods to be cached
c8y settings update cache.methods "GET PUT POST"
```

</CodeExample>

### cache.path: string

Location of the cache directory. If the path does not exist it was be created when the first cached item is created.

Defaults to:

* Linux/MacOS: `{tmp}/go-c8y-cli-cache`, where `{tmp}` is set to `$TMPDIR` if not empty otherwise `/tmp`
* Windows: `{tmp}/go-c8y-cli-cache`, where `{tmp}` is set to the first non-empty value of `%TMP%`, `%TEMP%`, `%USERPROFILE%`

:::note
If go-c8y-cli is being by multiple users on the same computer/server and you need to prevent users from accessing the local cached files, then you should change the default location of the cache, and apply the appropriate folder permissions to prevent other users from accessing the response from other users.
:::

### ci: boolean

Enable CI mode where all prompts will be disabled. If set to `true` it will override/ignore the following settings:

* `mode.enablecreate`
* `mode.enabledelete`
* `mode.enableupdate`

:::note
As it's name suggests, CI mode is ideally used for CI/CD pipelines only. The settings can be easily updated by setting a single environment variable:

```bash
export C8Y_SETTINGS_CI=true
```
:::

### encryption.cachepassphrase: boolean

Cache your encryption passphrase in an environment variable.

### encryption.enabled: boolean

Enable session encryption where your session password and/or token will be stored encrypted in the session file.

You may need to run `set-session` on your session again to apply the session encryption.

### includeall.delayms: int

Delay in milliseconds used internally by go-c8y-cli when iterating through the pages when using the `includeAll` parameter.

### includeall.pagesize: int

Page size to be used when using the `includeAll` parameter. This value will override any `pageSize` parameter.

### logger.hidesensitive: boolean

Hide sensitive information such usernames, passwords and tokens when using `dry` to make it easier to share the documentation with other people whilst keeping your passwords safe.

### mode.confirmation: string

Type of HTTP methods which will cause the user to be prompted.

You can customize this to only prompt when editing and deleting items using:

<CodeExample>

```bash
c8y settings update mode.confirmation "PUT DELETE"
```

</CodeExample>


### mode.enablecreate: boolean

Enable `POST` commands. If set to `false` then all `POST` related commands will return an error.

### mode.enabledelete: boolean

Enable `DELETE` commands. If set to `false` then all `DELETE` related commands will return an error.

### mode.enableupdate: boolean

Enable `UPDATE` commands. If set to `false` then all `UPDATE` related commands will return an error.

### session.defaultusername: string

Default username which is used when creating a new command via `c8y sessions create`

### storage.storepassword: boolean

Enable storage of your password in your session file. Disable if you do not want to store sensitive information to file. However this means you will be prompted for your password when you select a session via `set-session`.


### storage.storetoken: boolean

Enable storage of your OAUTH token in your session file. The token automatically generated when using `set-session` if the tenant uses OAUTH2_INTERNAL.

### template.path: string

Directory where the templates are located which will be available via tab completion and reference by name.

### template.customPaths: string

Additional directories where templates are located. Multiple paths can be specified by using a `:` delimiter. Environment variables can also be referenced using `$VARNAME`.

### views.columnmaxwidth: int

Maximum width of columns when using the table view


### views.columnminwidth: int

Minimum width of columns when using the table view

### views.columnMinWidthEmptyValue: int

Minimum column width for cells where the first row includes an non-existent or empty value when using the table view. This value is typically larger than the minimum width so that other rows which might have an non-empty value will still be readable (and not truncated so much)

### views.columnpadding: int

Column padding to add when calculating the column widths based on the content.

### views.commonpaths: []string

Array of directories where view files should be search for. 

:::note
It is recommended to only use this setting in your global configuration file, so it can be reused between multiple sessions.
:::

### views.custompaths: []string

Array of directories where view files should be search for

:::note
It is recommended to only use this setting in your session file if you want a custom view which is only valid for a single tenant.
:::

### views.rowmode: string

Row rendering mode. Accepts `truncate`, `wrap` or `overflow`. `truncated` is the default setting, however if there is an invalid settings, then `overflow` will be used.

In `wrap` mode, row separators will also be included to better visually delimit the table cells.
