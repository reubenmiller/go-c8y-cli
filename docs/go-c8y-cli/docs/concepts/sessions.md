---
layout: default
category: Concepts
title: Sessions
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeExample from '@site/src/components/CodeExample';

A session holds the current Cumulocity settings and authentication to use for each command. For example, as session will contain the Cumulocity platform address, tenant, username and password. All of these settings are required in order to send REST requests to the platform.

To use a session, you must first "activate" it, which will then start the "login" process where it uses the given session details and checks what login actions are required. For instance, if you have Two-Factor-Authentication (TFA) enabled, then you'll be prompted for your TFA code.

The session activate process is fairly complex due to the number of different login options supported by Cumulocity, however **go-c8y-cli** tries to simply the process by performing the following steps automatically, and only prompting for information where required:

* Check for missing session information (e.g. password), then prompt if necessary
* Check if session data is encrypted, then prompt for encryption passphrase to decrypt it
* Check preferred login method (e.g. BASIC_AUTH, OAUTH2_INTERNAL) and run the corresponding steps to then fetch a token if necessary
* Check if Two-Factor-Authentication setup is required, and prompt the user to configure it via a QR code
* Check if Two-Factor-Authentication code is requested, then prompt the user for it

The output of a successfully activate session is a set of environment variables that need to be set on the shell. Due to OS security restrictions, no subprocess (**c8y** in this can) can modify environment variables of the parent process (the **shell**), so instead shell helper function is required to evaluate the output of the session activation which in turn sets the required environment variables. The main motivation for using environment variables is to also allow other Cumulocity tooling to be able to access the required session information to run their own requests without having to know anything about session files and decryption etc.

:::caution
SSO (Single Sign On) is not currently supported due to a security mechanism on the platform side which makes any cli implementation impossible.

If you are an SSO user, then you will have to do one of the following before you can use go-c8y-cli:

* Create a dedicated local user in Cumulocity (via the Administration -> Users page)
* Create a service user via the Application User interface (though only for advanced users whom already have a username/password, see the [example](../../cli/c8y/microservices/serviceusers/c8y_microservices_serviceusers_create/#examples))

If you are using a local Cumulocity user, it recommended that you use TFA (Two-Factor Authentication) and use the "OAI-Secure" preferred login mode, which enables the usage of tokens (e.g. Bearer Authorization header), all of which is supported out of the box by go-c8y-cli.
:::

## Create a new session

The following command can be used to create a new session where you'll be prompted for the required information.

1. Create a new session

    <CodeExample>

    ```bash
    c8y sessions create
    ```

    </CodeExample>

    You will be prompted for the session details like the host, username and password.

    :::tip
    By default, go-c8y-cli will encrypt sensitive information such as passwords and tokens. The data is encrypted by using user-prompted passphrase and salt (in the form of a file under `$HOME/.cumulocity/.key` by default)
    :::

2. Activate the session (using the *set-session* shell helper)

    <CodeExample>

    ```bash
    set-session
    ```

    </CodeExample>

### Create a session for a host with self-signed certificates

If your Cumulocity host is using a self-signed certificate (which is usually the case if you are using the Cumulocity Edge), then you may want to use the `allowInsecure` flag. This option ignores the SSL verification. It should only be used if you trust the host and you have no option available (e.g. importing the missing CA certificate to the operating system)!

<CodeExample>

```bash
c8y sessions create --allowInsecure
```

</CodeExample>

Alternatively you can allow insecure mode (i.e. disable SSL verification) for a single command by using the `--insecure` global flag.

<CodeExample>

```bash
c8y devices list --insecure
```

</CodeExample>

## Activate a session (interactive)

A helper, called `set-session`, is provided to set the session interactively by providing the user a list of configured sessions. The user will be prompted to select one of the listed sessions (if there is more than one session, otherwise the single session will be automatically selected).

:::note
On some terminals, you need to hold `shift+ArrowKey` to navigate the list of sessions. 

Alternatively, VIM style shortcuts "j" (down) and "k" (up) keys can be also used for navigation. Though this does not work when your using the interact search. 
:::


<CodeExample>

```bash
set-session
```

</CodeExample>

:::tip
You can also provide additional flags to set session, as these are passed directly to the underlying `c8y sessions login` command.
:::

## Activate a session (manually)

If you can't use the `set-session` shell helper, then you can run the command directly along with the associated shell built-in functions which will set the required environment variables. Each shell has a slightly different way of doing this, so make sure you select the method appropriate to your shell

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Bash', value: 'bash', },
    { label: 'Zsh', value: 'zsh', },
    { label: 'Shell (posix)', value: 'sh', },
    { label: 'Fish', value: 'fish', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
eval "$(c8y session login)"
```

</TabItem>
<TabItem value="zsh">

```bash
eval "$(c8y session login)"
```

</TabItem>
<TabItem value="sh">

```bash
eval "$(c8y session login)"
```

</TabItem>
<TabItem value="fish">

```bash
c8y sessions login | source
```

</TabItem>
<TabItem value="powershell">

```powershell
c8y sessions login | Out-String | Invoke-Expression
```

</TabItem>
</Tabs>


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

* `C8Y_HOST` (example "https://cumulocity.com")
* `C8Y_TENANT` (example "myTenant")
* `C8Y_USER`
* `C8Y_PASSWORD`
* `CI`


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

The **go-c8y-cli** provides a large number of commands which can be potentially destructive if used incorrectly. Therefore all commands which create, update and/or delete data are disabled by default, to protect against accidental usage, this means by default only commands which just "read" information from the platform are enabled.

Whilst this may seem like overkill to new users, especially if you're just starting out your journey with Cumulocity and are in exploration/development mode, the value of this feature will become apparent once you start dealing with more than one Cumulocity instance and start going into production. For example, once you have multiple sessions, it is easy for users to accidentally forget where they are, and run a destructive command like "delete all devices" where the user thought they were in a dev instance, however they forgot they had switched to a production instance. This feature is not meant to be a substitute for proper platform side role/permission control, however it can safe powerusers against innocent mistakes.

The types of commands which are enabled or disabled is controlled by the **session mode**. The session mode can be controlled in the following ways:

* Set the default session mode when creating a new session (e.g. if you know it is a production instance, then set it to "prod")
* Override the default session mode when activating a session (e.g. if you know you need to do some additional actions on a session that is not normally allowed by default)
* Change the mode for a single command

Below shows the examples of how the session mode can be controlled in difference situations.

### Set the session mode when activating a session

<CodeExample transform="true">

```bash
set-session --mode dev
```

</CodeExample>

If you don't know the exact values you want, then you can ask to be prompted to select from a list of modes using:

<CodeExample transform="true">

```bash
set-session --mode prompt
```

</CodeExample>

### Enabling create/update/delete command temporarily

When the commands are disabled in the session settings, they can be temporarily activated on the console by using the following command:

<CodeExample transform="false">

```bash
export C8Y_MODE=dev
```

```powershell
$env:C8Y_MODE = "dev"
```

```powershell
$env:C8Y_MODE = "dev"
```

</CodeExample>

The commands will remain enabled until the next time you call `set-session`.


### Change the default mode of an existing session

Once you've activate a session, you can change it's default session mode by running the following command (though you will need to reload the environment to activate)

<CodeExample transform="true">

```bash
c8y settings update mode dev
```

</CodeExample>

Then reload your session by running `set-session` again.


### Change session mode for a single command

If you try to run a command which is not enabled, you will be asked if this was intentional and be asked to confirm the action. You'll also be presented with the session's information to help you make that decision to show if you're in the correct environment or not.

Alternatively, you can avoid the user prompt by using the `sessionMode` global flag.

<CodeExample transform="true">

```bash
c8y inventory create --name hello --sessionMode dev
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
        "session.mode": "prod"
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
eval "$( c8y sessions login --shell=auto )"

# Set a session to an already known json path.
eval "$( c8y sessions login --shell=auto --session "/my/path/session.json" )"
```

</TabItem>
<TabItem value="fish">

```bash
c8y sessions login --shell=auto | source

# Set a session to an already known json path.
c8y sessions login --shell=auto --session "/my/path/session.json" | source
```

</TabItem>
<TabItem value="powershell">

```powershell
c8y sessions login --shell=auto | Out-String | Invoke-Expression

# Set a session to an already known json path.
c8y sessions login --shell=auto --session "/my/path/session.json" | Out-String | Invoke-Expression
```

</TabItem>
</Tabs>

:::info
`set-session` is a small helper function (for each supported shell) which wraps the call to `c8y sessions login` and sets the returned environment variables which are then read by subsequent calls to %%c8y%%.
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
# Using VSCode
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

