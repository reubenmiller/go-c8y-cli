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

:::note
On some terminals, you need to hold `shift+ArrowKey` to navigate the list of sessions. 

Alternatively, VIM style shortcuts "j" (down) and "k" (up) keys can be also used for navigation. Though this does not work when your using the interact search. 
:::

:::caution
`set-session` is not provided by `c8y` itself, and it is installed automatically for you if you following the [installation guide](/docs/installation/shell-installation)
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

All of the values in the sessions file, can also be overridden using environment variables. The   [configuration settings](/docs/configuration/settings) pages details how to modify them.


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
    "$schema": "https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/v2/tools/schema/session.schema.json",
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

## Advanced session commands

The following shows some other ways how to manage sessions.

### Setting a session without the helper

If you have not installed the go-c8y-cli addons, then you cannot use the `set-session` helper function.

You need to switch sessions by calling the c8y binary directly and evaluating the environment variables returned by the command. Each shell handles this slightly different.

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Bash/zsh', value: 'bash', },
    { label: 'Fish', value: 'fish', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
eval $( c8y sessions set --shell=auto )

# Set a session to an already known json path.
eval $( c8y sessions set --shell=auto --session "/my/path/session.json" )
```

</TabItem>
<TabItem value="fish">

```bash
c8y sessions set --shell=auto | source

# Set a session to an already known json path.
c8y sessions set --shell=auto --session "/my/path/session.json" | source
```

</TabItem>
<TabItem value="powershell">

```powershell
c8y sessions set --shell=auto | Out-String | Invoke-Expression

# Set a session to an already known json path.
c8y sessions set --shell=auto --session "/my/path/session.json" | Out-String | Invoke-Expression
```

</TabItem>
</Tabs>

:::info
`set-session` is a small helper function (for each supported shell) which wraps the call to `c8y session set` and sets the returned environment variables which are then read by subsequent calls to `c8y`.
:::

### Switching session for a single command

A single command can be redirected to use another session by using the `session <name>` parameter. A full file path can be provided or just the file name for a file located in session directory defined by the `C8Y_SESSION_HOME` environment variable.

<CodeExample>

```bash
c8y devices list --session myothersession.json
```

</CodeExample>

### Creating a session file manually

Session files can also be manually created as each session is just a file. The file needs to be placed on the configured session folder. 

You can check your where your session home folder is by running

<CodeExample transform="false">

```bash
c8y settings list --select session.home

# or if you want to write it to a variable
myhome=$( c8y settings list --select session.home --output csv )
```

```powershell
c8y settings list --select session.home

# or if you want to write it to a variable
$myhome = c8y settings list --select session.home --output csv
```

</CodeExample>

```bash title="Output"
| session.home                            |
|-----------------------------------------|
| /workspaces/go-c8y-cli/.cumulocity      |
```

Then you can create the a file using your preferred text editor (i.e. vim, nano, VSCode)

<CodeExample>

```bash
# Using vim
vim ~/.cumulocity/my-manual-file.json
```

```powershell
# using VSCode
code ~/.cumulocity/my-manual-file.json
```

</CodeExample>

```powershell title="file: ~/.cumulocity/session1.json"
{
    "host": "example01.cumulocity.eu-latest.com",
    "username": "hello.user@example.com",
    "password": "mys3cureP4assw!rd",
}
```

### Cloning an existing session

You can easily clone (copy) your activated session by running. The mode can also be changed when cloning the session.

<CodeExample transform="false">

```bash
c8y sessions clone --newName "customer-qual" --type "qual"
```

</CodeExample>

```bash title="Output"
✓ Cloned session file to /workspaces/go-c8y-cli/.cumulocity/customer-qual.json
```

Then switch to the cloned session

<CodeExample transform="false">

```bash
set-session "customer-qual"
```

</CodeExample>

:::tip
Cloning an existing session is convenient when you a group of tenants (i.e. dev, qual, prod) where the settings only differ by url and your password. So you can just clone it and then edit the file manually in your preferred text editor.

Remember it is best practice to use different passwords for different sessions!
:::

### Changing mode of the existing session until next set-session

The mode can be updated on the existing profile using
#### Change mode until next set-session

If you installed the addons then there are some helpers to set the mode temporarily until the session is changed.

<CodeExample transform="false">

```bash
set-c8ymode-dev
```

```powershell
set-c8ymode-dev
```

</CodeExample>

#### Changing the mode permanently

<CodeExample transform="false">

```bash
c8y settings update mode dev
```

</CodeExample>
