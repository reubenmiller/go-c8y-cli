---
layout: default
category: Concepts
title: AI agents (MCP)
---

go-c8y-cli includes a [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server and an agent skill so AI agents, such as Claude Code, can work with a Cumulocity tenant while you stay in control of what they can do.

## How it works

Your MCP client starts `c8y mcp serve` and communicates with it over stdin and stdout, so you only register it once and don't need to keep a terminal open. As the server has no terminal, you select and unlock sessions in your browser:

* **Session**: Without `--session`, the first tool call opens a session picker in your browser. When you select a session, the server logs in straight away and shows the result on the page, so you can retry or pick another session if the login fails. The agent's tool call waits until the login succeeds (for up to 3 minutes). If you cancel, the agent is asked to check with you, and the picker opens again on the next tool call.
* **Encrypted sessions**: If you use a pin entry command, such as a password manager or keychain helper, it provides the passphrase when you select the session. Configure it once with a setting:

    ```sh
    c8y settings update pinEntry "touchie get"
    ```

    Without a pin entry command, the page asks for the passphrase of encrypted sessions and checks it before logging in. This also applies to a session given with `--session`.
* **Two-factor authentication**: If the login needs a code, the page asks for the code from your authenticator app, unless the session contains a TOTP secret. Setting up two-factor authentication for the first time needs a terminal, so run `c8y sessions login` once in a terminal first.
* **Switching sessions**: Run the server's `switch_session` prompt, for example `/mcp__c8y__switch_session` in Claude Code. Only you can run it, not the agent. The session picker opens in the browser, and the agent's tool calls wait until you have selected and logged in to a session. Commands which are already running finish with the previous session.

The session picker only listens on `127.0.0.1` while it is open, and can only be opened with the one-time link which the server opens in your browser. The link is never shown to the agent. The passphrase and the two-factor code are passed to the login through a pipe, not environment variables, and aren't kept.

The server gives the agent three tools:

| Tool | Description |
|------|-------------|
| `c8y_session` | Show the tenant, user, allowed mode and command policy. No secrets are returned |
| `c8y_read` | Run a c8y command which only reads data, e.g. `devices list --select id,name` |
| `c8y_write` | Run a c8y command which creates, updates or deletes data. Marked as destructive, so clients ask for permission |

Each tool call runs a normal `c8y` command, so the existing go-c8y-cli features apply:

* **Session**: selected by you in the browser. The agent can't switch it, and can't use `--session`, `--sessionUsername`, `--sessionPassword`, `--host` or `--proxy`. The agent is told to check the active session with `c8y_session`, as you can switch it at any time.
* **Credentials**: the commands only receive the token from the login and a copy of the session file without secrets. The password and encryption passphrase are never passed to the commands, except the password for sessions which use basic auth (no token). The token and password are passed through a pipe, not environment variables, as other programs running as your user can read the environment of a process. On Windows, environment variables are still used.
* **Session files**: the commands can't read or write files in the session home folder (`C8Y_SESSION_HOME`, e.g. `~/.cumulocity`) or the session file, so the agent can't read secrets with `--template`, `--data`, `--templateVars`, `--file`, or a jsonnet `import`. Use `--denyPath` to protect other folders, such as `~/.ssh`. Paths in your [`files.deniedPaths`](/docs/configuration/settings#filesdeniedpaths-string) setting are denied as well.
* **Token renewal**: when the token is about to expire, the server logs in again. If the passphrase or a two-factor code is needed, the page opens again in your browser.
* **Session mode**: by default, the agent uses the mode of each session (`settings.session.mode`), so it can only read data in a `prod` session and can also delete data in a `dev` session. A session without a mode is read only. The mode set with `--mode` (or in the policy file) is a maximum for every session, limited by each session's own mode. For example, with `--mode qual`, the agent can create and update data in a `dev` session, but only read data after you switch to a `prod` session. Read commands always run in `prod` (read-only) mode, and CI environment variables are ignored.
* **Dry run**: commands with `--dry` are treated as read commands, and the approval prompt shows the dry run output of the requests.
* **Activity log**: each command is recorded in the [activity log](/docs/concepts/activitylog) with the message `mcp:c8y_read` or `mcp:c8y_write`.
* **Sensitive information**: always hidden in log output.

## Setup

Register the server in your MCP client (once).

**Claude Code**

```sh
# select the session in the browser on the first tool call
claude mcp add --scope user c8y -- c8y mcp serve

# use a specific session, and at most create and update data
claude mcp add --scope user c8y -- c8y mcp serve --session my-dev-tenant --mode qual
```

**Other clients (Claude Desktop, VS Code, Cursor)**

```json
{
  "mcpServers": {
    "c8y": {
      "command": "c8y",
      "args": ["mcp", "serve"]
    }
  }
}
```

The server opens the session picker with a desktop notification (macOS, and Linux with `notify-send`). Use `--noNotify` to turn the notification off.

### Automation and CI

In CI, set the credentials as environment variables in the MCP client configuration. The server then uses them without a session picker:

```json
{
  "mcpServers": {
    "c8y": {
      "command": "c8y",
      "args": ["mcp", "serve", "--mode", "qual"],
      "env": {
        "C8Y_HOST": "https://example.cumulocity.com",
        "C8Y_TENANT": "t12345",
        "C8Y_USER": "ci-user",
        "C8Y_PASSWORD": "<password>"
      }
    }
  }
}
```

You can also provide `C8Y_TOKEN` instead of a password. In this case, the commands receive the credentials from the server's environment.

## Controlling what the agent can do

| Flag | Description |
|------|-------------|
| `--mode` | Maximum mode: `prod` (read only), `qual` (create and update) or `dev` (create, update and delete). Default: the session's mode, or `prod` if the session has none |
| `--allow` | Only allow commands matching the pattern. Can be used multiple times |
| `--deny` | Deny commands matching the pattern (wins over `--allow`). Can be used multiple times |
| `--read` | Treat additional commands as read-only, e.g. extension commands |
| `--denyPath` | Local file or folder the commands can't read, e.g. `~/.ssh`. The session home folder and the session file are always denied. Can be used multiple times |
| `--approval` | How write commands are approved: `auto` (default), `elicit`, `client` or `none` |
| `--policy` | Load the policy from a json, yaml or toml file. Flags override values in the file |
| `--timeout`, `--maxTimeout` | Default and maximum timeout per command |
| `--maxOutput` | Maximum number of bytes returned to the agent per command |
| `--noDefaultDeny` | Don't include the default deny list |

Patterns match command paths word by word and match all sub commands, for example `devices` matches `devices list` and `devices children list`, and `* delete` matches every delete command at the second level.

By default the following commands are denied, because they expose credentials, change the local configuration or open local network services: `mcp`, `sessions`, `settings`, `cli`, `alias`, `completion`, `extensions`, `microservices getBootstrapUser`, `deviceregistration getCredentials`, `remoteaccess connect run` and `remoteaccess server`.

### Troubleshooting devices

For troubleshooting, the agent can run shell commands on a device. This is disabled by default. Add `--allowDeviceShell` to the server's arguments to enable it:

```sh
claude mcp add --scope user c8y -- c8y mcp serve --mode qual --allowDeviceShell
```

Both ways are write commands, so they also need `--mode qual` or `--mode dev`, and are approved like any other write command:

| Method | Example | Requirements |
|--------|---------|--------------|
| ssh via Cumulocity remote access | `remoteaccess connect ssh --device 12345 -- systemctl status` | A remote access (passthrough) configuration for the device's ssh port. The ssh login uses your ssh keys or ssh agent, and the user from `--user` or the `remoteaccess.sshuser` setting |
| `c8y_Command` operation | `operations create --device 12345 --data "c8y_Command.text=df -h" \| operations wait --status SUCCESSFUL --status FAILED` | The device supports the `c8y_Command` operation |

For `remoteaccess connect ssh`, the MCP server only allows running a single command:

* The command to run on the device must be given after `--`. Interactive sessions are rejected.
* `--listen` and `-L`/`--port-forward` are rejected, as they open local network listeners.
* The command is always treated as a write command, also with `--dry`, as it has no dry run mode. For the same reason, the approval prompt doesn't show a dry run of the requests, but notes that the command runs on the device.
* The command runs in ssh batch mode (`--batch`): only your ssh keys are used, and ssh fails instead of waiting for a password or a key passphrase. Without `--user` or the `remoteaccess.sshuser` setting, your local user name is tried first, then `root`.
* The ssh connection only exists while the command runs. When the timeout is reached, the command and its ssh process are stopped.

`remoteaccess connect run` stays denied, as it runs a command on your machine, and `remoteaccess server` as it opens a local listener.

A shell command on a device can change anything the ssh or agent user is allowed to change, regardless of the mode. To confirm each command yourself, add `--confirmDeviceShell`:

```sh
claude mcp add --scope user c8y -- c8y mcp serve --mode qual --allowDeviceShell --confirmDeviceShell
```

The server then shows a dialog with the command and the tenant after the normal approval (macOS, and Linux with `zenity` or `kdialog`). If no dialog is available, the command is rejected. Unlike the normal approval, the MCP client can't answer it. Deny is the default, and the command is rejected if you don't answer within 2 minutes.

Without `--allowDeviceShell`, `remoteaccess connect ssh` is rejected, and so are `operations create` and `c8y api` commands whose arguments or input contain `c8y_Command`. The check of the operations is best effort: a template can assemble the operation in other ways. To make sure the agent can't send any operation to a device, also deny `operations create`:

```sh
claude mcp add --scope user c8y -- c8y mcp serve --mode qual --deny "operations create"
```

### Approval

| Value | Behavior |
|-------|----------|
| `auto` | Ask you through MCP elicitation if the client supports it, otherwise rely on the client's tool permission prompt |
| `elicit` | Always ask you through MCP elicitation. Write commands are rejected if the client doesn't support it |
| `client` | Rely on the client's tool permission prompt |
| `none` | No approval. Only use this for trusted automation |

The elicitation prompt shows the command, the tenant and the requests from a dry run.

With MCP clients which use protocol version 2026-07-28 or later, the question is returned with the tool result and the client calls the tool again with your answer. With older clients, the server sends an elicitation request instead. In both cases, an approval is only valid once, and only for the exact command, input and session you were asked about. If you switch the session before answering, the agent has to run the command again.

The approval of write commands is handled by the MCP client: either the client shows the elicitation prompt, or its tool permission prompt. It protects against unwanted actions of the agent, but not against a modified MCP client, which can answer the prompts itself.

### Policy file

```yaml
# c8y-mcp-policy.yaml
mode: qual
approval: elicit
allow: [devices, inventory, alarms, events, measurements, operations, api]
deny: ["* delete"]
readCommands: ["mycompany list"]
denyPaths: ["~/.ssh", "~/.aws"]
timeout: 2m
maxOutputBytes: 100000
```

```sh
claude mcp add --scope user c8y -- c8y mcp serve --policy /path/to/c8y-mcp-policy.yaml
```

### Limits

* Read commands are detected by their name (`get*`, `list*`, `find*`, `count*`, `exists*`) and by the HTTP method for `c8y api`. Even if a command is detected incorrectly, it still runs in `prod` mode, so the session mode check rejects any write.
* The policy controls the `c8y` commands run through the MCP server. It doesn't stop an agent which also has shell access from running `c8y` or reading files directly. Use your client's permission settings to control shell access.
* Commands can read local files outside the denied folders, for example with `--template`, `--data` or jsonnet imports, and the output is returned to the agent. Add folders with secrets to `--denyPath`, and don't rely on the MCP server alone to keep local files away from an agent.
* Use encrypted sessions. A session file outside the session home folder is only protected while it is the active session.
* Commands are run without a shell. Only c8y commands joined with ` | ` are supported.
* The session picker opens in a browser on the machine which runs the server. On a remote machine (e.g. over ssh), use `--session` with a session which can log in without user input, or environment variables.
* `remoteaccess connect ssh` can't ask for a password or a key passphrase, as it runs in batch mode. Use an ssh key without a passphrase, or one which is loaded in your ssh agent.

## Agent skill

The [go-c8y-cli skill](https://github.com/reubenmiller/go-c8y-cli/tree/v2/skills/go-c8y-cli) teaches agents how to use go-c8y-cli: discovering commands, keeping output small, previewing writes with `--dry`, and using pipelines. It works with both the MCP tools and the `c8y` command in a shell.

To install it for Claude Code, copy the `skills/go-c8y-cli` folder to `~/.claude/skills/` (all projects) or `.claude/skills/` (one project).
