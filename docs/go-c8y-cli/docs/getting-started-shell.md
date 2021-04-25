---
id: gettingstarted
title: Getting started
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

Before you can send any requests to Cumulocity you need to configure the Cumulocity session which has the details which Cumulocity platform and authentication should be used for each of the commands/requests. This process only needs to be done once.


## Create a new session

<Tabs
    groupId="shell-types"
    defaultValue="bash"
    values={[
        { label: 'Shell', value: 'bash', },
        { label: 'PowerShell', value: 'powershell', },
    ]
    }>
<TabItem value="bash">

```bash
c8y sessions create \
    --host "https://mytenant.eu-latest.cumulocity.com" \
    --username "myUser@me.com" \
    --type dev
```

</TabItem>
<TabItem value="powershell">

```powershell
New-Session `
    -Host "https://mytenant.eu-latest.cumulocity.com" \
    -Username "myUser@me.com" \
    -Type dev
```

</TabItem>

</Tabs>

You will be prompted for your password. Alternatively you can also enter the password using the `password` parameter.

You may also provide a more meaningful session name by using the `name` parameter.

The `type` parameter indicates what kind of tenant you are using which directly controls which commands are enabled by default, and which are disabled. `dev` will enable all commands, whereas `prod` only allows GET commands.

Alternatively you can create a session by creating a json file in your `~/.cumulocity` folder.

### Example: Manually create a session file

```powershell title="file: ~/.cumulocity/session1.json"
{
    "host": "example01.cumulocity.eu-latest.com",
    "username": "hello.user@example.com",
    "password": "mys3cureP4assw!rd",
}
```


## Activate the session using the interactive session selector

```bash
set-session
```

The list of sessions can be filtered by adding additional filter terms when calling `set-session`. If only 1 session is found, then it will be automatically selected without the user having to confirm the selection.

```bash
set-session eu example
```

## Test your credentials by getting your current user information from the platform

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y currentuser get

# or list devices
c8y devices list
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-CurrentUser

# or list devices
Get-DeviceCollection
```

</TabItem>
</Tabs>


:::note
If your credentials are incorrect, then you can update the session file stored in the `~/.cumulocity` directory
:::


## Switching sessions

The sessions can be changed again by using the interactive session selector

```bash
set-session
```

Alternatively you can switch sessions by calling the c8y binary directly (without the set-session helper)

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


### Switching session for a single command

A single command can be redirected to use another session by using the `session <name>` parameter. A full file path can be provided or just the file name for a file located in session directory defined by the `C8Y_SESSION_HOME` environment variable.

```bash
c8y devices list --session myothersession.json
```
