---
id: gettingstarted
title: Getting started
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeExample from '@site/src/components/CodeExample';

Before you can send any requests to Cumulocity you need to configure the Cumulocity session which has the details which Cumulocity platform and authentication should be used for each of the commands/requests. This process only needs to be done once.

:::caution
SSO (Single Sign On) is not currently supported due to a security mechanism on the platform side which makes any cli implementation impossible.

If you are an SSO user, then you will have to do one of the following before you can use go-c8y-cli:

* Create a dedicated local user in Cumulocity (via the Administration -> Users page)
* Create a service user via the Application User interface (though only for advanced users whom already have a username/password, see the [example](../cli/c8y/microservices/serviceusers/c8y_microservices_serviceusers_create/#examples))

If you are using a local Cumulocity user, it recommended that you use TFA (Two-Factor Authentication) and use the "OAI-Secure" preferred login mode, which enables the usage of tokens (e.g. Bearer Authorization header), all of which is supported out of the box by go-c8y-cli.
:::

## Basics

### Creating a new session

First, you'll need to create a new session which contains your credentials to interact with Cumulocity. However, creating a session does not activate it, so after the session is created, then activate it by following the instructions in the next session.

<CodeExample>

```bash
c8y sessions create
```

</CodeExample>

You will be prompted the session information including url, username and password. Alternatively, you can provide any of the settings via flags.

### Activating a session (interactive)

To use **go-c8y-cli** with Cumulocity, you'll need to activate a session which tells all subsequent **c8y** commands which Cumulocity tenant the API calls should be sent to, and what authorization should be used.

:::info
The `set-session` is a shell function included with **go-c8y-cli** which just sets the session information as environment variables which are accessed by other **c8y** commands which are executed in the same shell.

If you get an unknown command error when running `set-session`, it means you probably didn't install [installation guide](/docs/installation/shell-installation). However if you're still having trouble, then try to [activate a session manually](#activate-a-session-without-set-session).
:::


<CodeExample>

```bash
set-session
```

</CodeExample>

Or you can also change the default session mode by using the `--mode` flag which allows you to override the session's default.

<CodeExample>

```bash
set-session --mode dev
```

</CodeExample>


#### set-session filtering

The list of sessions can be filtered by adding additional filter terms when calling `set-session`, where the filter terms will be used to match against any of the session's properties, e.g. url, username, tenant etc.

```bash
set-session eu example
```

:::note
If only 1 session is found, then it will be automatically selected without the user having to confirm the selection
:::

### Checking the active session

You can check which session is activated by using:

<CodeExample>

```bash
c8y sessions get
```

</CodeExample>


## Advanced

### Activate a session (without set-session)

A session can be activated without the shell helper function `set-session`. It is useful if you don't want to install the plugin script and are just using the binary directly.

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Bash', value: 'bash', },
    { label: 'Shell (posix)', value: 'sh', },
    { label: 'Zsh', value: 'zsh', },
    { label: 'Fish', value: 'fish', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
eval "$(c8y sessions login --shell bash)"
```

</TabItem>
<TabItem value="sh">

```bash
eval "$(c8y sessions login --shell sh)"
```

</TabItem>
<TabItem value="zsh">

```bash
eval "$(c8y sessions login --shell zsh)"
```

</TabItem>
<TabItem value="fish">

```bash
c8y sessions login --shell fish | source
```

</TabItem>
<TabItem value="powershell">

```powershell
c8y sessions login --shell powershell | out-string | Invoke-Expression
```

</TabItem>
</Tabs>
