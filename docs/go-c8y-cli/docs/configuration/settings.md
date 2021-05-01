---
layout: default
category: Configuration
title: Settings
---

PSc8y and c8y supports a modern approach to configuration. It allows you to control settings via configuration files and/or environment settings.


#### Order of processing

The settings are read and applied in the following order:

1. Read `settings.<json|yaml>` inside current working directory
2. Read `settings.<json|yaml>` inside the `C8Y_SESSION_HOME` env variable
3. Read session file path in the `C8Y_SESSION` env variable
4. Read setting from env variable (if defined)

The value last set will be used by `c8y`.

#### Options

The following table lists the available settings, the environment variable equivalent and description.

| Name | Environment Variable | Description |
|------|----------------------|-------------|
| default.pageSize | `C8Y_SETTINGS_DEFAULTS_PAGESIZE` | Default page size |
| includeAll.pageSize | `C8Y_SETTINGS_INCLUDEALL_PAGESIZE` | Default page size when using the includeAll parameter |
| includeAll.delayMS | `C8Y_SETTINGS_INCLUDEALL_DELAYMS` | Delay between fetching the next page when using the includeAll parameter |
| template.path | `C8Y_SETTINGS_TEMPLATE_PATH` | Path / Folder where the templates are located. If the user gives a template name (without path), then a matching filename will be search for in this folder |
| ci | `C8Y_SETTINGS_CI` | Enable CI/CD mode where any command restrictions will be disabled |
| mode.enableCreate | `C8Y_SETTINGS_MODE_ENABLECREATE` | Enable/disable create commands |
| mode.enableUpdate | `C8Y_SETTINGS_MODE_ENABLEUPDATE` | Enable/disable update commands |
| mode.enableDelete | `C8Y_SETTINGS_MODE_ENABLEDELETE` | Enable/disable delete commands |

### Example: Set global defaults to use in each c8y session

Global settings can be controlled by creating a `settings.json` file in the `~/.cumulocity` folder, and adding the following contents:

File: *~/.cumulocity/settings.json*

```json
{
  "$schema": "https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/master/tools/schema/session.schema.json",
  "settings": {
      "default.pageSize": 2000,
      "includeAll.pageSize": 100,
      "includeAll.delayMS": 500
  }
}
```

The same settings can also be added to your session file, so that you can override the defaults defined in the `settings.json` file.

Below shows an example of a session (json) file with a `settings` section, where the default pageSize is set to 50.

File: *~/.cumulocity/my-session01.json*

```json
{
  "$schema": "https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/master/tools/schema/session.schema.json",
  "host": "https://example.zz-latest.cumulocity.com",
  "tenant": "t12345",
  "username": "hans@example.com",
  "password": "h4n$gRu8er",
  "description": "",
  "settings": {
      "default.pageSize": 50
  }
}
```

**Notes**

* If the same settings exists in the `settings.json` and session file, then the value in the session file will be used
* The `$schema` property is provided to enable tab completion of the properties when editing in VS Code (or another editor which supports the json schema lookups)

### Environment variables

Some settings cannot be defined in the session or settings files as they control the behaviour related to such files.

The following is a list of available environment variables which control how c8y and PSc8y interacts with activating, searching and displaying c8y sessions.

| Environment Variable | Description |
|----------------------|-------------|
| C8Y_SESSION | Path to the Cumulocity session file to be used. |
| C8Y_SESSION_HOME | Path where the session files and settings are located. Defaults to `~/.cumulocity` if it is not set. |
| C8Y_SETTINGS_LOGGER_HIDESENSITIVE | Control whether sensitive session information is logged to the console or not. |
| C8Y_JSONNET_DEBUG | Display debugging information for jsonnet templates (if used) |
| C8Y_DISABLE_ENFORCE_ENCODING | (PowerShell only) Disable enforcement of UTF8 encoding on the console. If UTF8 encoding is disabled it will cause encoding problems if non-ascii characters are used! |


### Environment variable details and examples   

#### C8Y_SESSION

Path to the Cumulocity session file to be used.

If it exists when the PSc8y PowerShell module is loaded, then the session will be loaded automatically.

---

#### C8Y_SESSION_HOME

By default the `$HOME/.cumulocity` directory is used to store the Cumulocity session files. A custom session home folder can be specified by setting the `C8Y_SESSION_HOME` to a folder.

Use a custom folder where the Cumulocity Session files should be kept and searched through.

---

#### C8Y_SETTINGS_LOGGER_HIDESENSITIVE

Control whether sensitive session information is logged to the console or not. When set to `true`, then session information such as `tenant`, `username`, `password`, `basic auth header` will be obfuscated. If the setting is not present, then the session information will be shown (except for clear-text passwords).

When using the `PSc8y` PowerShell module, you can change the settings using the `Set-ClientConsoleSetting` cmdlet:

**Example**

```powershell
Set-ClientConsoleSetting -HideSensitive

# or enable it again by using
Set-ClientConsoleSetting -ShowSensitive
```

On bash, you can configured it by setting the environment variable:

```bash
export C8Y_SETTINGS_LOGGER_HIDESENSITIVE=true
```
