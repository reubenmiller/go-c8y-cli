---
category: mcp
title: c8y mcp serve
---
Start an MCP server for AI agents

### Synopsis

Start a Model Context Protocol (MCP) server which gives AI agents controlled access to a
Cumulocity tenant using c8y commands.

By default you run the server in your own terminal. It logs in using the normal session login,
so encrypted sessions, two-factor authentication and SSO work as usual, then listens on a
unix socket which only your user can access. If no session is set (via --session or C8Y_SESSION),
you are prompted to select one. MCP clients connect to it with "c8y mcp connect".
The token is renewed by logging in again in the terminal when it is about to expire.

You can switch to another session while the server is running by typing "s" and Enter in the
server terminal. Connected clients use the new session for new commands, so the MCP client only
needs to be registered once. Type "i" to show the current session and "h" for help.

Each new client must be approved in the server terminal, which shows the process that is connecting
and the processes that started it. This stops other programs running as your user from using the
server without you noticing, as the approval can't be given by the client. The client connects
straight away, and its tool calls wait until it is approved, so it doesn't time out. The terminal
bell and a desktop notification draw your attention to the prompt (disable with --noNotify).
One answer applies to all connections from the same client which are waiting, and answering "a"
also allows later connections from that client (e.g. reconnects), until it hasn't connected for
8 hours. "a" is only offered when the connecting process is this c8y binary (c8y mcp connect).
The socket must be in a directory which is owned by your user and not writable by other users.

With --stdio, the MCP client starts the server itself and communicates over stdin/stdout. The
session must then work without user input, e.g. by setting C8Y_HOST, C8Y_TENANT, C8Y_USER and
C8Y_PASSWORD or C8Y_TOKEN in the MCP client configuration, or C8Y_PASSPHRASE for encrypted
sessions. This mode is intended for automation and CI.

You control what the agent can do:

* Session: selected by you in the server terminal. The agent can't switch the session, host or
  credentials. In socket mode, commands only receive the token, not the password or passphrase.
* Mode (prod, qual, dev): the maximum mode the agent can use. It can only narrow the session's own
  mode, never widen it, also after switching the session. Read-only commands always run in prod
  (read-only) mode.
* Allow and deny lists: which commands the agent can use, e.g. "devices", "inventory *", "* delete".
* Device shell: with --allowDeviceShell (and qual or dev mode), the agent can run single shell
  commands on a device for troubleshooting, via "remoteaccess connect ssh --device \<id> -- \<command>"
  (if remote access is configured) or c8y_Command operations. Add --confirmDeviceShell to confirm
  each command in a dialog or the server terminal, which the MCP client can't answer.
* Approval: how write commands are approved:
    auto    ask via MCP elicitation if the client supports it, otherwise rely on the client's tool permission prompt
    elicit  always ask via MCP elicitation and reject writes if the client does not support it
    client  rely on the client's tool permission prompt (c8y_write is marked as destructive)
    none    no approval (only for trusted automation)

Each command run by the agent is recorded in the c8y activity log with the message "mcp:\<tool>".

The policy can also be loaded from a file (json, yaml or toml):

    mode: qual
    approval: elicit
    allow: [devices, inventory, alarms, events, measurements, operations, api]
    deny: ["* delete"]
    readCommands: ["mycompany list"]
    timeout: 2m
    maxOutputBytes: 100000


```
c8y mcp serve [flags]
```

### Examples

```
$ c8y mcp serve
Select a session interactively, log in and start the server in the current terminal (read only)

$ c8y mcp serve --session my-dev-tenant --mode qual
Log in to the session and start the server in the current terminal (allowing create and update commands)

$ claude mcp add --scope user c8y -- c8y mcp connect
Add the server to Claude Code (run once). Claude Code starts the relay, which connects to the running server

$ c8y mcp serve --session dev --mode dev --deny "users" --deny "tenants" --approval elicit
Allow all actions (including delete) except for user and tenant management, and always ask the user to approve writes

$ claude mcp add c8y -e C8Y_HOST=https://example.cumulocity.com -e C8Y_USER=ci -e C8Y_PASSWORD=... -- c8y mcp serve --stdio
Let the MCP client start the server, using credentials from environment variables

```

### Options

```
      --allow stringArray                 Only allow commands matching this pattern, e.g. devices, "inventory *". Can be used multiple times
      --allowDeviceShell[=true|false]     Allow the agent to run shell commands on devices (remoteaccess connect ssh and c8y_Command operations). Needs --mode qual or dev
      --approval string                   How write commands are approved: auto, elicit, client or none. Default: auto
      --confirmDeviceShell[=true|false]   Ask for confirmation of each command run on a device in a dialog or the server terminal, in addition to the normal approval
      --deny stringArray                  Deny commands matching this pattern, e.g. users, "* delete". Can be used multiple times
  -h, --help[=true|false]                 help for serve
      --maxOutput int                     Maximum number of output bytes returned to the agent per command. Default: 100000
      --maxTimeout duration               Maximum timeout the agent can request per command. Default: 10m
      --mode string                       Maximum mode the agent can use: prod (read only), qual (create, update) or dev (create, update, delete). Default: prod
      --noDefaultDeny[=true|false]        Don't include the default deny list: mcp, sessions, settings, cli, alias, completion, extensions, microservices getBootstrapUser, deviceregistration getCredentials, remoteaccess connect run, remoteaccess server
      --noNotify[=true|false]             Don't show a desktop notification when a client needs to be approved
      --policy string                     Policy file (json, yaml or toml). Flags override the values in the file
      --read stringArray                  Additional commands to treat as read-only, e.g. extension commands. Can be used multiple times
      --socket string                     Unix socket path to listen on, e.g. to run several servers at the same time. Defaults to server.sock in the mcp folder of the session home directory
      --stdio[=true|false]                Communicate over stdin/stdout (the MCP client starts the server). The session must work without user input
      --timeout duration                  Default timeout per command. Default: 2m
```

### Options inherited from parent commands

```
      --abortOnErrors int                Abort batch when reaching specified number of errors (default 10)
      --allowEmptyPipe[=true|false]      Don't fail when piped input is empty (stdin)
      --cache[=true|false]               Enable cached responses
      --cacheBodyPaths strings           Cache should limit hashing of selected paths in the json body. Empty indicates all values
      --cacheTTL string                  Cache time-to-live (TTL) as a duration, i.e. 60s, 2m (default "60s")
  -c, --compact[=true|false]             Compact instead of pretty-printed output when using json output. Pretty print is the default if output is the terminal
      --confirm[=true|false]             Prompt for confirmation
      --confirmText string               Custom confirmation text
      --currentPage int                  Current page which should be returned
      --customQueryParam strings         add custom URL query parameters. i.e. --customQueryParam 'withCustomOption=true,myOtherOption=myvalue'
      --debug[=true|false]               Set very verbose log messages
      --delay string                     delay after each request, i.e. 5ms, 1.2s (default "0ms")
      --delayBefore string               delay before each request, i.e. 5ms, 1.2s (default "0ms")
      --dry[=true|false]                 Dry run. Don't send any data to the server
      --dryFormat string                 Dry run output format. i.e. json, dump, markdown or curl (default "markdown")
      --examples[=true|false]            Show examples for the current command
      --filter stringArray               Apply a client side filter to response before returning it to the user
      --flatten[=true|false]             flatten json output by replacing nested json properties with properties where their names are represented by dot notation
  -f, --force[=true|false]               Do not prompt for confirmation. Ignored when using --confirm
  -H, --header strings                   custom headers. i.e. --header "Accept: value, AnotherHeader: myvalue"
      --includeAll[=true|false]          Include all results by iterating through each page
  -k, --insecure[=true|false]            Allow insecure server connections when using SSL
  -l, --logMessage string                Add custom message to the activity log
      --maxJobs int                      Maximum number of jobs. 0 = unlimited (use with caution!)
      --noAccept[=true|false]            Ignore Accept header will remove the Accept header from requests, however PUT and POST requests will only see the effect
      --noCache[=true|false]             Force disabling of cached responses (overwrites cache setting)
  -M, --noColor[=true|false]             Don't use colors when displaying log entries on the console
      --noLog[=true|false]               Disables the activity log for the current command
      --noProgress[=true|false]          Disable progress bars
      --noProxy[=true|false]             Ignore the proxy settings
  -n, --nullInput[=true|false]           Don't read the input (stdin). Useful if using in shell for/while loops
  -o, --output string                    Output format i.e. table, json, csv, csvheader (default "table")
      --outputFile string                Save JSON output to file (after select/view)
      --outputFileRaw string             Save raw response to file (before select/view)
      --outputTemplate string            jsonnet template to apply to the output
  -p, --pageSize int                     Maximum results per page (default 5)
      --progress[=true|false]            Show progress bar. This will also disable any other verbose output
      --proxy string                     Proxy setting, i.e. http://10.0.0.1:8080
  -r, --raw[=true|false]                 Show raw response. This mode will force output=json and view=off
      --retries int                      Max number of attempts when a failed http call is encountered (default 3)
      --select stringArray               Comma separated list of properties to return. wildcards and globstar accepted, i.e. --select 'id,name,type,**.serialNumber'
      --session string                   Session configuration
      --sessionMode string               Override default session mode for a single command which would normally be disabled
  -P, --sessionPassword string           Override session password
  -U, --sessionUsername string           Override session username. i.e. peter or t1234/peter (with tenant)
      --silentExit[=true|false]          Silent status codes do not affect the exit code
      --silentStatusCodes string         Status codes which will not print out an error message
      --totalPages int                   Total number of pages to get
  -v, --verbose[=true|false]             Verbose logging
      --view string                      Use views when displaying data on the terminal. Disable using --view off (default "auto")
      --withError[=true|false]           Errors will be printed on stdout instead of stderr
      --withTotalElements[=true|false]   Request Cumulocity to include the total elements in the response statistics under .statistics.totalElements (introduced in 10.13)
  -t, --withTotalPages[=true|false]      Request Cumulocity to include the total pages in the response statistics under .statistics.totalPages
      --workers int                      Number of workers (default 1)
```

