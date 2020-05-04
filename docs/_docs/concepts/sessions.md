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
    --host "https://mytenant.eu-latest.cumulocity.com" \
    --tenant "mytenant" \
    --username "https://mytenant.eu-latest.cumulocity.com"
```

You will be prompted for the password if the `--password` argument is not used.

##### Powershell

```powershell
New-Session
```

### Activate a session (interactive)

A helper is provided to set the session interactively by providing the user a list of configured sessions. The user will be prompted to select one of the listed sessions.

##### Bash

```sh
export C8Y_SESSION=$( c8y sessions list )
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
