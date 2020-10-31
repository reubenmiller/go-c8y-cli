---
layout: default
category: Concepts
title: Sessions
---

A session holds the current Cumulocity settings and authentication to use for each command. For example, as session will contain the Cumulocity platform address, tenant, username and password. All of these settings are required in order to send REST requests to the platform.

The active session is controlled via an environment variable `C8Y_SESSION`. The environment variable just points to a JSON file which contains the Cumulocity settings.

A user can add any number of session that that want. They are then free to switch sessions at any time they want. For users with only one session, you could set the `C8Y_SESSION` environment variable to a file, however for users with multiple sessions, it is recommended to always to set the session when starting out.

### Create a new session

##### Bash

```bash
c8y sessions create \
    --host "mytenant.eu-latest.cumulocity.com"
```

You will be prompted for the username and password.

##### Powershell

```powershell
New-Session
```

### Activate a session (interactive)

A helper is provided to set the session interactively by providing the user a list of configured sessions. The user will be prompted to select one of the listed sessions.

##### Bash

Assuming that you have already loaded the `c8y.profile.sh` helper.

```sh
set-session
```

##### zsh

Assuming that you have already loaded the `c8y.plugin.zsh` plugin.

```sh
set-session
```

##### Powershell

```sh
Set-Session
```

### Active a session (manual)

If you wish to manage your sessions manually, then you can set the `C8Y_SESSION` environment variable to an existing json session file.

For example:

###### Bash

```sh
export C8Y_SESSION=~/.cumulocity/my-settings01.json
```

###### Powershell

```powershell
$env:C8Y_SESSION = "~/.cumulocity/my-settings01.json"
```

### Session file format

A Cumulocity session file is simply a json file. Here is an example of the contents:

```json
{
  "host": "https://mytenant.eu-latest.cumulocity.com",
  "tenant": "mytenant",
  "username": "myTestUser",
  "password": "sUp3rs3curEpassW0r5",
  "description": "My test Cumulocity tenant",
  "useTenantPrefix": false
}
```

You can edit this file manually to update your password, or to fix any mistakes made when entering your details.

The session file is read every time when a command is executed, so any changes will be read next time you use a command.

Optional properties:

By default the tenant prefix will be used in the basic authentication, however if you do not what this behavior, then you can add the `use`

For example:

```sh
{
  "host": "https://mytenant.eu-latest.cumulocity.com",
  "tenant": "myTenant",
  "username": "myTestUser",
  "password": "sUp3rs3curEpassW0r5",
  "description": "My test Cumulocity tenant",
  "useTenantPrefix": false
}
```

All of the values in the sessions file, can also be overridden using environment variables when either the `--useEnv` switch or the `C8Y_USE_ENVIRONMENT=true` variable is set. The following is the mapping between the properties in the session json file to their environment equivalents.

| Setting name | Environment Variable |
|--------------|----------------------|
| useTenantPrefix | C8Y_USETENANTPREFIX |


### Continuous Integration usage (environment variables)

Alternatively, the Cumulocity session can be controlled purely by environment variables.

Firstly, the `C8Y_USE_ENVIRONMENT` environment needs to be set to `true` to activate this mode.

Then the Cumulocity settings can be set by the following environment variables.

* C8Y_HOST (example "https://cumulocity.com")
* C8Y_TENANT (example "myTenant")
* C8Y_USER
* C8Y_PASSWORD
* C8Y_SETTINGS_CI


### Switching sessions for a single command

If you only need to set a session for a single session, then you can use the global `--session` argument. The name of the session should be the name of the file stored under your `~/.cumulocity/` folder (with or without the .json extension).

You can set the `C8Y_SESSION_HOME` environment variable to control where the sessions should be stored.

###### Bash

```sh
c8y devices list --session myother.tenant
```

###### Powershell

```powershell
Get-DeviceCollection -Session myother.tenant
```

### Protection against accidental data loss

The c8y cli tool provides a large number of commands which can be potentially destructive if used incorrectly. Therefore all commands which create, update and/or delete data are disabled by default, to protect against accidental usage.

The commands can be enabled per session or in global settings, however explicitily enabling it per session is the perferred method.

Ideally for production sessions, the settings should be left disabled to protect yourself against accidental data loss, especially if you have a large number of tenants, and are constantly switching between them.

The following shows how the functions can be controlled via session settings:

*File: mysession.json*

```json
{
  "settings": {
    "mode.enableCreate": false,
    "mode.enableUpdate": false,
    "mode.enableDelete": false
  }
}
```

#### Enabling create/update/delete command temporarily

When the commands are disabled in the session settings, they can be temporarily activated on the console by using the following command:

**PowerShell**

```powershell
Set-ClientConsoleSetting -EnableCreateCommands -EnableUpdateCommands -EnableDeleteCommands
```

**Bash/zsh**

```sh
export C8Y_SETTINGS_MODE_ENABLECREATE=true
export C8Y_SETTINGS_MODE_ENABLEUPDATE=true
export C8Y_SETTINGS_MODE_ENABLEDELETE=true
```

The commands will remain enabled until the next time you call `set-session`.

Afterwards you can disable them again using:

**PowerShell**

```powershell
Set-ClientConsoleSetting -DisableCommands
```

**Bash/zsh**

```sh
unset C8Y_SETTINGS_MODE_ENABLECREATE
unset C8Y_SETTINGS_MODE_ENABLEUPDATE
unset C8Y_SETTINGS_MODE_ENABLEDELETE
```

#### CI/CD

When used in a CI/CD environment, all commands (create/update/delete) can be enabled by setting the `settings.ci` property to `true`.

The setting can also be set via an environment variable:

**Bash/zsh**

```sh
export C8Y_SETTINGS_CI=true
```

**PowerShell**

```sh
$env:C8Y_SETTINGS_CI=true
```

Or alternatively, using setting it via the session file:

*File: mysession.json*

```json
{
  "settings": {
    "ci": true
  }
}
```

### Storing passwords in the Cumulocity session

Encryption is used to stored sensitive session information such as passwords and authorization cookies. The user will be prompted for the passphrase if one is not already set, when activating a session.

The passphrase should be something that is sufficiently complex and should not be stored on disk.

When the user sets the passphrase, a key file will be created within the Cumulocity session home folder, `.key`. This file will be used as a reference when comparing your passphrase to keep the passphrase constant across different sessions.

### Loss of passphrase (encryption key)

If you forget your passphrase then all of the encrypted passwords will be unuseble.

In such an event, then you need to remove the `.key` file within the Cumulocity session folder, and you will be prompted to re-enter your password when the session is re-activated using `set-session`.

**PowerShell**

```Powershell
Remove-Item ~/.cumulocity/.key
```

**Bash/zsh**
```sh
rm ~/.cumulocity/.key
```

## Manually settings password via the file


```json
{
    "$schema": "https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/master/tools/schema/session.schema.json",
    "host": "https://example.cumulocity.com",
    "tenant": "t12345",
    "username": "myuser@iot-user.com",
    "password": "{encrypted}65cd99f96f9fe681be286d6e573061053afac353faeb5b1220352ab57456f3ee852fa9078ead3846c982caad6c4dfd3be6fd0a9aba",
    "description": "",
    "settings": {
        "mode.enableUpdate": true
    }
}
```

#### Updating passwords

Passwords can still be set as plain text in the session files, however the next time that you switch to the session using `set-session`, the `password` field will be encrypted. An field is marked as encrypted by starting with text `{encrypted}` followed by the encrypted string.
