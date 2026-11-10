---
layout: default
category: Concepts
title: AI agents (MCP)
---

go-c8y-cli includes a [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server and an agent skill so AI agents, such as Claude Code, can work with a Cumulocity tenant while you stay in control of what they can do.

## How it works

You run `c8y mcp serve` in a terminal. It logs in to the session using the normal session login, so encrypted sessions, two-factor authentication and SSO work as usual. If you don't set a session with `--session` or `C8Y_SESSION`, you are asked to select one, the same as with `set-session`. The server then listens on a unix socket which only your user can access. You can switch to another session at any time in the server's terminal, without changing the MCP client configuration.

Your MCP client starts `c8y mcp connect`, which forwards the client's messages to the server's socket. The relay doesn't use any credentials. If the server isn't running, the agent is told to ask you to start it.

Each new client must be approved in the server's terminal. The prompt shows the process which is connecting and the processes which started it, for example `c8y mcp connect` started by `claude`. It stops programs running as your user from using the logged-in session without you noticing. It doesn't stop a program which targets your setup on purpose, as such a program could also change your MCP client's configuration or replace the `c8y` binary.

```text
your terminal                                  MCP client (e.g. Claude Code)
c8y mcp serve                          ◄────── c8y mcp connect
  (session selection, login prompts,     unix socket, only accessible by your user
   connection approval, activity log)
```

The server gives the agent three tools:

| Tool | Description |
|------|-------------|
| `c8y_session` | Show the tenant, user, allowed mode and command policy. No secrets are returned |
| `c8y_read` | Run a c8y command which only reads data, e.g. `devices list --select id,name` |
| `c8y_write` | Run a c8y command which creates, updates or deletes data. Marked as destructive, so clients ask for permission |

Each tool call runs a normal `c8y` command, so the existing go-c8y-cli features apply:

* **Session**: selected by you in the server's terminal. The agent can't switch it, and can't use `--session`, `--sessionUsername`, `--sessionPassword`, `--host` or `--proxy`. The agent is told to check the active session with `c8y_session`, as you can switch it at any time.
* **Credentials**: the commands only receive the token from the login and a copy of the session file without secrets. The password and encryption passphrase stay in the server. Sessions which use basic auth (no token) also pass the password to the commands. The token and password are passed through a pipe, not environment variables, as other programs running as your user can read the environment of a process. On Windows, environment variables are still used.
* **Token renewal**: when the token is about to expire, the server asks you to log in again in its terminal.
* **Session mode**: read commands always run in `prod` (read-only) mode. Write commands run in the mode you set with `--mode`. It can only narrow the session's own mode, never widen it, also after switching to another session, and CI environment variables are ignored.
* **Dry run**: commands with `--dry` are treated as read commands, and the approval prompt shows the dry run output of the requests.
* **Activity log**: each command is recorded in the [activity log](/docs/concepts/activitylog) with the message `mcp:c8y_read` or `mcp:c8y_write`, and shown in the server's terminal.
* **Sensitive information**: always hidden in log output.

## Setup

1. Start the server in a terminal and log in:

    ```sh
    # select a session interactively (the agent uses the session's mode)
    c8y mcp serve

    # use a specific session
    c8y mcp serve --session my-tenant

    # at most create and update data, even if the session allows deleting
    c8y mcp serve --session my-dev-tenant --mode qual
    ```

2. Register the relay in your MCP client (once). The server prints the command to use.

    **Claude Code**

    ```sh
    claude mcp add --scope user c8y -- c8y mcp connect
    ```

    **Other clients (Claude Desktop, VS Code, Cursor)**

    ```json
    {
      "mcpServers": {
        "c8y": {
          "command": "c8y",
          "args": ["mcp", "connect"]
        }
      }
    }
    ```

3. Approve the connection in the server's terminal when your MCP client starts. Check that the process is the one you expect, then answer:

    ```text
    A new MCP client is connecting:
      c8y mcp connect
          pid 52124, /opt/homebrew/bin/c8y
      started by: claude
          pid 88512, /Users/me/.local/bin/claude
    Allow this client to use the MCP server? (a = also allow later connections from claude, pid 88512) [y/N/a]
    ```

    | Answer | Behavior |
    |--------|----------|
    | `y` | Allow the connections which are waiting. You are asked again when the client reconnects |
    | `a` | Also allow later connections from the same client process (e.g. reconnects). A new client process, such as a restarted Claude Code, is asked again, and so is a client which hasn't connected for 8 hours |
    | `n` | Reject the connections which are waiting |

    The client connects straight away and doesn't have to wait for your answer, so it never times out. Until you answer, the agent's tool calls wait for up to 10 seconds, and then the agent is asked to remind you to approve the client. The server also rings the terminal bell and shows a desktop notification (macOS, and Linux with `notify-send`). Use `--noNotify` to turn the notification off.

    MCP clients can open more than one connection, for example when reconnecting. One answer applies to all connections from the same client which are waiting, and the prompt shows how many there are. If the client disconnects before you answer, the prompt is cancelled.

    `a` is only offered when the connecting process is the `c8y` binary of the server, i.e. the `c8y mcp connect` relay. It allows all relays started by that client process, including ones started by its hooks or shell commands.

4. Keep the terminal open while the agent works. Press Ctrl+C to stop the server, which immediately removes the agent's access.

### Switching sessions

Type a command in the server's terminal and press Enter:

| Command | Description |
|---------|-------------|
| `s` | Switch to another session. Select the session from the list and log in. Press Ctrl+C in the list to cancel |
| `i` | Show the current session |
| `h` | Show the available commands |

Connected clients use the new session for new commands, so you don't need to reconnect the MCP client. Commands which are already running finish with the previous session. If the login to the new session fails, the server keeps using the previous session.

By default, the agent uses the mode of each session (`settings.session.mode`), so it can only read data in a `prod` session and can also delete data in a `dev` session. A session without a mode is read only. The mode set with `--mode` (or in the policy file) is a maximum for every session, limited by each session's own mode. For example, with `--mode qual`, the agent can create and update data in a `dev` session, but only read data after you switch to a `prod` session.

### Socket

The socket is `server.sock` in the `mcp` folder of the session home directory, so the server and the relay find each other no matter which session is active. To run several servers at the same time, for example for two MCP clients working with different tenants, use `--socket <path>` on both commands. The folder containing the socket must be owned by your user and must not be writable by other users, so shared folders such as `/tmp` are rejected. `c8y mcp connect` also checks that the server runs as your user before connecting.

### Without a terminal (stdio)

With `--stdio`, the MCP client starts the server itself and communicates over stdin and stdout, so you don't need to keep a terminal open. Register the server instead of the relay:

```sh
claude mcp add --scope user c8y -- c8y mcp serve --stdio
```

As the server has no terminal, it asks for input in other ways:

* **Session**: Without `--session`, the first tool call opens a session picker in your browser. When you select a session, the server logs in straight away and shows the result on the page, so you can retry or pick another session if the login fails. The agent's tool call waits until the login succeeds (for up to 3 minutes). If you cancel, the agent is asked to check with you, and the picker opens again on the next tool call.
* **Encrypted sessions**: If you use a pin entry command, such as a password manager or keychain helper, it provides the passphrase when you select the session. Configure it once with a setting instead of an environment variable:

    ```sh
    c8y settings update pinEntry "touchie get"
    ```

    Without a pin entry command, the page asks for the passphrase of encrypted sessions and checks it before logging in. This also applies to a session given with `--session`. The passphrase is only passed to the login and isn't kept, so if the session's token expires, the page opens again to ask for it.
* **Two-factor authentication**: The session must contain a TOTP secret, as there is nowhere to enter a code.

To switch to another session, run the server's `switch_session` prompt, for example `/mcp__c8y__switch_session` in Claude Code. Only you can run it, not the agent. The session picker opens in the browser, and the agent's tool calls wait until you select a session. The new session applies to new commands. As in the terminal, the agent uses the new session's mode, limited by `--mode` if set.

The session picker only listens on `127.0.0.1` while it is open, and can only be opened with the one-time link which the server opens in the browser. The link is never shown to the agent.

#### Automation and CI

In CI, set the credentials as environment variables in the MCP client configuration. The server then uses them without a session picker:

```json
{
  "mcpServers": {
    "c8y": {
      "command": "c8y",
      "args": ["mcp", "serve", "--stdio", "--mode", "qual"],
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

You can also provide `C8Y_TOKEN` instead of a password. In this mode, the commands receive the credentials from the server's environment.

## Controlling what the agent can do

| Flag | Description |
|------|-------------|
| `--mode` | Maximum mode: `prod` (read only), `qual` (create and update) or `dev` (create, update and delete). Default: the session's mode, or `prod` if the session has none |
| `--allow` | Only allow commands matching the pattern. Can be used multiple times |
| `--deny` | Deny commands matching the pattern (wins over `--allow`). Can be used multiple times |
| `--read` | Treat additional commands as read-only, e.g. extension commands |
| `--approval` | How write commands are approved: `auto` (default), `elicit`, `client` or `none` |
| `--policy` | Load the policy from a json, yaml or toml file. Flags override values in the file |
| `--timeout`, `--maxTimeout` | Default and maximum timeout per command |
| `--maxOutput` | Maximum number of bytes returned to the agent per command |
| `--noDefaultDeny` | Don't include the default deny list |

Patterns match command paths word by word and match all sub commands, for example `devices` matches `devices list` and `devices children list`, and `* delete` matches every delete command at the second level.

By default the following commands are denied, because they expose credentials, change the local configuration or open local network services: `mcp`, `sessions`, `settings`, `cli`, `alias`, `completion`, `extensions`, `microservices getBootstrapUser`, `deviceregistration getCredentials`, `remoteaccess connect run` and `remoteaccess server`.

### Troubleshooting devices

For troubleshooting, the agent can run shell commands on a device. This is disabled by default. Start the server with `--allowDeviceShell` to enable it:

```sh
c8y mcp serve --mode qual --allowDeviceShell
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
* The ssh connection only exists while the command runs. When the timeout is reached, the command and its ssh process are stopped.

`remoteaccess connect run` stays denied, as it runs a command on your machine, and `remoteaccess server` as it opens a local listener.

A shell command on a device can change anything the ssh or agent user is allowed to change, regardless of the mode. To confirm each command yourself, add `--confirmDeviceShell`:

```sh
c8y mcp serve --mode qual --allowDeviceShell --confirmDeviceShell
```

The server then shows a dialog with the command and the tenant after the normal approval (macOS, and Linux with `zenity` or `kdialog`), or asks in its terminal if no dialog is available. Unlike the normal approval, the MCP client can't answer it. Deny is the default, and the command is rejected if you don't answer within 2 minutes.

Without `--allowDeviceShell`, `remoteaccess connect ssh` is rejected, and so are `operations create` and `c8y api` commands whose arguments or input contain `c8y_Command`. The check of the operations is best effort: a template can assemble the operation in other ways. To make sure the agent can't send any operation to a device, also deny `operations create`:

```sh
c8y mcp serve --mode qual --deny "operations create"
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

The approval of write commands is handled by the MCP client: either the client shows the elicitation prompt, or its tool permission prompt. It protects against unwanted actions of the agent, but not against a program which connects to the socket directly and answers the prompts itself. The connection approval in the server's terminal makes that harder.

### Policy file

```yaml
# c8y-mcp-policy.yaml
mode: qual
approval: elicit
allow: [devices, inventory, alarms, events, measurements, operations, api]
deny: ["* delete"]
readCommands: ["mycompany list"]
timeout: 2m
maxOutputBytes: 100000
```

```sh
c8y mcp serve --session my-dev-tenant --policy ./c8y-mcp-policy.yaml
```

### Limits

* Read commands are detected by their name (`get*`, `list*`, `find*`, `count*`, `exists*`) and by the HTTP method for `c8y api`. Even if a command is detected incorrectly, it still runs in `prod` mode, so the session mode check rejects any write.
* The policy controls the `c8y` commands run through the MCP server. It doesn't stop an agent which also has shell access from running `c8y` or reading files directly. Use your client's permission settings to control shell access.
* Commands can read local files, for example with `--template`, `--data` or jsonnet imports, and the output is returned to the agent. Don't rely on the MCP server alone to keep local files away from an agent.
* Commands are run without a shell. Only c8y commands joined with ` | ` are supported.
* The connection approval shows the names and paths of the connecting processes. A malicious program can use a misleading name, so check the executable path, and reject connections which you didn't expect.
* On Windows, the socket folder check is skipped and the connecting process isn't shown.
* `remoteaccess connect ssh` can't ask for a password or a key passphrase, as the command doesn't run in a terminal. Use an ssh key which is loaded in your ssh agent.

## Agent skill

The [go-c8y-cli skill](https://github.com/reubenmiller/go-c8y-cli/tree/v2/skills/go-c8y-cli) teaches agents how to use go-c8y-cli: discovering commands, keeping output small, previewing writes with `--dry`, and using pipelines. It works with both the MCP tools and the `c8y` command in a shell.

To install it for Claude Code, copy the `skills/go-c8y-cli` folder to `~/.claude/skills/` (all projects) or `.claude/skills/` (one project).
