---
id: gettingstarted
title: Getting started
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeExample from '@site/src/components/CodeExample';

Before you can send any requests to Cumulocity you need to configure the Cumulocity session which has the details which Cumulocity platform and authentication should be used for each of the commands/requests. This process only needs to be done once.


## Basics

### Creating a new session

<CodeExample>

```bash
c8y sessions create \
    --host "https://mytenant.eu-latest.cumulocity.com" \
    --username "myUser@me.com" \
    --type dev
```

</CodeExample>

You will be prompted for your password. Alternatively you can also enter the password using the `password` parameter.

You may also provide a more meaningful session name by using the `name` parameter.

The `type` parameter indicates what kind of tenant you are using which controls which commands are enabled/disabled by default. `dev` will enable all commands, whereas `prod` only allows GET commands.

### Activating a session (interactive)

<CodeExample>

```bash
set-session
```

</CodeExample>

:::info
If you get an unknown command error when running `set-session`, it means you probably didn't install the the addons via the [installation guide](installation/shell-installation)
:::

```text title="Output"
➜ set-session 
Use arrow keys (holding shift) to navigate ↓ ↑ → ←  and / toggles search
? Select a Cumulocity Session: 
  ▶ #01 json dev-poc                                  http://my-dev-tenant.example.com (t1111/rmiller-dev01)
    #02 json customer1-qual                           http://dev.customer-domain.com (t2222/myuser01)
    #03 json customer1-prod                           http://qual.customer-domain.com (t3333/myuser01)

--------- Details ----------
File:            /workspaces/go-c8y-cli/.cumulocity/dev-poc.json
Host:            http://my-dev-tenant.example.com
Tenant:          t1111
Username:        rmiller-dev01
```

The list of sessions can be filtered by adding additional filter terms when calling `set-session`.

```bash
set-session eu example
```

:::note
If only 1 session is found, then it will be automatically selected without the user having to confirm the selection
:::

### Validating a session

Once a session has been activated, you can check if everything works as expected by getting your user's details associated to your session you configured.

<CodeExample>

```bash
c8y currentuser get

# or list devices
c8y devices list
```

</CodeExample>

:::note
If your credentials are incorrect, then you can update the session file

```sh
c8y sessions get --select path
```
:::

## Advanced

### Activate a session (without set-session)

A session can be activated without the shell helper function `set-session`. It is useful if you don't want to install the plugin script and are just using the binary directly.

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Bash', value: 'bash', },
    { label: 'Zsh', value: 'zsh', },
    { label: 'Fish', value: 'fish', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
eval $(c8y sessions set --shell bash)
```

</TabItem>
<TabItem value="zsh">

```bash
eval $(c8y sessions set --shell zsh)
```

</TabItem>
<TabItem value="fish">

```bash
c8y sessions set --shell fish | source
```

</TabItem>
<TabItem value="powershell">

```powershell
c8y sessions set --shell powershell | out-string | Invoke-Expression
```

</TabItem>
</Tabs>
