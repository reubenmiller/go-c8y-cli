---
layout: default
category: Concepts
title: Sessions
---

import CodeExample from '@site/src/components/CodeExample';

A session holds the current Cumulocity settings and authentication to use for each command. For example, as session will contain the Cumulocity platform address, tenant, username and password. All of these settings are required in order to send REST requests to the platform.

The active session is controlled via an environment variable `C8Y_SESSION`. The environment variable just points to a JSON file which contains the Cumulocity settings.

A user can add any number of session that that want. They are then free to switch sessions at any time they want. For users with only one session, you could set the `C8Y_SESSION` environment variable to a file, however for users with multiple sessions, it is recommended to always to set the session when starting out.

## Create a new session

<CodeExample>

```bash
c8y sessions create \
    --host "mytenant.eu-latest.cumulocity.com"
```

</CodeExample>

You will be prompted for the username and password.

## Activate a session (interactive)

A helper is provided to set the session interactively by providing the user a list of configured sessions. The user will be prompted to select one of the listed sessions.

Note: On MacOS, you need to hold "option"+Arrow keys to navigate the list of sessions. Otherwise the VIM style "j" (down) and "k" (up) keys can be also used for navigation

:::caution
`set-session` is not provided by `c8y` itself, and it is installed automatically for you if you following the installation instructions
:::

<CodeExample>

```bash
set-session
```

</CodeExample>

## Activate a session (manual)

If you wish to manage your sessions manually, then you can set the `C8Y_SESSION` environment variable to an existing json session file.

<CodeExample>

```bash
export C8Y_SESSION=~/.cumulocity/my-settings01.json
```

```powershell
$env:C8Y_SESSION = "~/.cumulocity/my-settings01.json"
```

</CodeExample>

### Session file format

A Cumulocity session file is simply a json file. Here is an example of the contents:

```json title="mysession.json"
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

By default the tenant prefix will be used in the basic authentication, however if you do not what this behavior, then you can add the `useTenantPrefix` property.

```bash
{
  "host": "https://mytenant.eu-latest.cumulocity.com",
  "tenant": "myTenant",
  "username": "myTestUser",
  "password": "sUp3rs3curEpassW0r5",
  "description": "My test Cumulocity tenant",
  "useTenantPrefix": false
}
```

All of the values in the sessions file, can also be overridden using environment variables. The following is the mapping between the properties in the session json file to their environment equivalents.

| Setting name | Environment Variable |
|--------------|----------------------|
| useTenantPrefix | C8Y_USETENANTPREFIX |


### Continuous Integration usage (environment variables)

Alternatively, the Cumulocity session can be controlled purely by environment variables.

Then the Cumulocity settings can be set by the following environment variables.

* C8Y_HOST (example "https://cumulocity.com")
* C8Y_TENANT (example "myTenant")
* C8Y_USER
* C8Y_PASSWORD
* C8Y_SETTINGS_CI


### Switching sessions for a single command

If you only need to set a session for a single session, then you can use the global `--session` argument. The name of the session should be the name of the file stored under your `~/.cumulocity/` folder (with or without the .json extension).

You can set the `C8Y_SESSION_HOME` environment variable to control where the sessions should be stored.

<CodeExample>

```bash
c8y devices list --session myother.tenant
```

```powershell
Get-DeviceCollection -Session myother.tenant
```

</CodeExample>

## Protection against accidental data loss

The c8y cli tool provides a large number of commands which can be potentially destructive if used incorrectly. Therefore all commands which create, update and/or delete data are disabled by default, to protect against accidental usage.

The commands can be enabled per session or in global settings, however explicitly enabling it per session is the preferred method.

Ideally for production sessions, the settings should be left disabled to protect yourself against accidental data loss, especially if you have a large number of tenants, and are constantly switching between them.

The following shows how the functions can be controlled via session settings:

```json title="mysession.json"
{
  "settings": {
    "mode.enableCreate": false,
    "mode.enableUpdate": false,
    "mode.enableDelete": false
  }
}
```

### Enabling create/update/delete command temporarily

When the commands are disabled in the session settings, they can be temporarily activated on the console by using the following command:

<CodeExample>

```bash
eval $( c8y settings update mode dev --shell auto )
```

```powershell
c8y settings update mode dev --shell auto | Out-String | Invoke-Expression
```

</CodeExample>


The commands will remain enabled until the next time you call `set-session`.

Afterwards you can disable them again using:

<CodeExample>

```bash
eval $( c8y settings update mode prod --shell auto )
```

```powershell
c8y settings update mode prod --shell auto | Out-String | Invoke-Expression
```

</CodeExample>

## Using encryption in Cumulocity session files

Encrypted password and cookies fields can be activated by adding the following fragment into the session file or your global `settings.json` file.

```json
{
  "settings": {
    "encryption": {
      "enabled": true
    }
  }
}
```

When enabled the "password", and "authorization.cookies" fields will be encrypted using a passphrase chosen by the user.
The passphrase should be something that is sufficiently complex and should not be stored on disk.

When the user sets the passphrase, a key file will be created within the Cumulocity session home folder, `.key`. This file will be used as a reference when comparing your passphrase to keep the passphrase constant across different sessions.

The user will be prompted for the passphrase if one is not already set, when activating a session.

### Loss of passphrase (encryption key)

If you forget your passphrase then all of the encrypted passwords will be unusable.

In such an event, then you need to remove the `.key` file within the Cumulocity session folder, and you will be prompted to re-enter your password when the session is re-activated using `set-session`.

<CodeExample>

```bash
rm ~/.cumulocity/.key
```

```powershell
Remove-Item ~/.cumulocity/.key
```

</CodeExample>

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

### Updating passwords

Passwords can still be set as plain text in the session files, however the next time that you switch to the session using `set-session`, the `password` field will be encrypted. An field is marked as encrypted by starting with text `{encrypted}` followed by the encrypted string.
