---
id: gettingstarted
title: Getting started
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeExample from '@site/src/components/CodeExample';

Before you can send any requests to Cumulocity you need to configure the Cumulocity session which has the details which Cumulocity platform and authentication should be used for each of the commands/requests. This process only needs to be done once.

:::tip
SSO users can authenticate via the browser-based **Authorization Code** flow (`--loginType BROWSER`) or the headless **Device Authorization** flow (`--loginType DEVICE`). See [Creating a session with SSO](#creating-a-session-with-sso) below.

If you need a non-SSO account, you can create a dedicated local user in Cumulocity (Administration → Users) or a [service user](../cli/c8y/microservices/serviceusers/c8y_microservices_serviceusers_create/#examples).
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

### Creating a session with SSO

go-c8y-cli supports two OAuth2-based SSO flows that don't require a username or password to be stored in the session file. The login type is saved in the session file so subsequent `set-session` calls use the same flow automatically.

**Browser Authorization Code Flow** — opens a browser at the tenant's SSO URL and waits for the redirect. Best for desktop use.

:::tip
The following is required for the Browser flow:
* The Cumulocity tenant must have **"Redirect to the user interface application"** enabled in its SSO configuration.
* Your SSO provider must allow the redirect URI `http://localhost:5001/callback` (customizable via `--browserCallback`).
:::

<CodeExample>

```bash
c8y sessions create \
  --host "https://example.cumulocity.com" \
  --loginType BROWSER
```

</CodeExample>

**Device Authorization Flow** — suited for headless environments. The CLI prints a URL and a short code; visit the URL in any browser to approve the login while the CLI polls for the result.

<CodeExample>

```bash
c8y sessions create \
  --host "https://example.cumulocity.com" \
  --loginType DEVICE
```

</CodeExample>

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
