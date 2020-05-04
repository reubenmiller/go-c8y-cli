---
layout: default
category: Configuration
title: Environment Variables
---

c8y and PSc8y can be configured using environment variables to control certain behaviors. The following is a list of available options.

### C8Y_SESSION

Path to the Cumulocity session file to be used.

If it exists when the PSc8y PowerShell module is loaded, then the session will be loaded automatically.

### C8Y_USE_ENVIRONMENT

When set to `on`, the Cumulocity session settings will be loaded from the following environment variables:

* C8Y_HOST (example "https://cumulocity.com")
* C8Y_TENANT (example "myTenant")
* C8Y_USER
* C8Y_PASSWORD

`C8Y_USE_ENVIRONMENT` will override the `C8Y_SESSION` environment variable.

```sh
# bash
export C8Y_USE_ENVIRONMENT=on

# PowerShell
$env:C8Y_USE_ENVIRONMENT = "on"
```

### C8Y_SESSION_HOME

By default the `$HOME/.cumulocity` directory is used to store the Cumulocity session files. A custom session home folder can be specified by setting the `C8Y_SESSION_HOME` to a folder.
Use a custom folder where the Cumulocity Session files should be keep and scanned.

### C8Y_LOGGER_HIDE_SENSITIVE

Control whether sensitive session information is logged to the console or not. When set to `true`, then session information such as `tenant`, `username`, `password`, `basic auth header` will be obfuscated. If the setting is not present, then the session information will be shown (except for clear-text passwords).

When using the `PSc8y` PowerShell module, you can change the settings using the `Set-CumulocityConsoleSetting` cmdlet:

```PowerShell
Set-CumulocityConsoleSetting -HideSensitive

# or enable it again by using
Set-CumulocityConsoleSetting -ShowSensitive
```

On bash, you can configured it by setting the environment variable:

```sh
export C8Y_LOGGER_HIDE_SENSITIVE=true
```

### PSC8Y_INSTALL_ON_IMPORT (PowerShell only)

On MacOS and Linux the c8y binary will be installed / copied to the /usr/local/bin directory when importing the PSc8y PowerShell module.

Example:

```PowerShell
export PSC8Y_INSTALL_ON_IMPORT=on
```

### C8Y_INSTALL_PATH (PowerShell only)

On MacOS and Linux, the install path where the c8y binary will be installed.
