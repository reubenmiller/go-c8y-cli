---
name: go-c8y-cli
description: Query and manage a Cumulocity tenant (devices, inventory, measurements, alarms, events, operations, users, applications) with go-c8y-cli, either through the `c8y` command in a shell or the go-c8y-cli MCP server tools (c8y_read, c8y_write, c8y_session). Use when the user asks about Cumulocity data, devices or tenants, or mentions c8y or go-c8y-cli.
---

# go-c8y-cli

go-c8y-cli (`c8y`) is a command line tool for the Cumulocity REST API. Commands follow the pattern
`c8y <resource> <verb> [flags]`, for example `c8y devices list --name "sensor*"`.

## Two ways to run commands

1. **MCP server tools (preferred when available)**. If the tools `c8y_read`, `c8y_write` and `c8y_session` exist, use them:
   - Call `c8y_session` once to see the tenant, the allowed mode and the command policy.
   - `c8y_read`: commands that only read data (get, list, find, count, `--help`, `--examples`, `--dry`).
   - `c8y_write`: commands that create, update or delete data. The user may be asked to approve each call.
   - Pass the command without a shell, e.g. `{"command": "devices list --name 'sensor*' --select id,name"}`.
     The leading `c8y` is optional. Pipes between c8y commands (` | `) work; `jq`, `grep`, `&&`, `;` and redirects do not.
   - Use the `stdin` field to pipe in data such as ids or JSON lines.
   - If a tool rejects a command because of the policy or mode, tell the user. Don't try to work around it.
   - The user selects the session; you can't change it. If they want another tenant, ask them to switch it
     (with the server's `switch_session` prompt), then call `c8y_session` again.
2. **Shell**. Run `c8y ...` directly. The session must already be active in the environment
   (`C8Y_SESSION` or `C8Y_HOST` plus credentials). Never run `c8y sessions set` or `c8y sessions login` yourself;
   they are interactive. Ask the user to activate a session instead.

## Rules

- **Don't guess flags.** Run `c8y <resource> <verb> --examples` (short) or `--help` (complete) first.
- **Preview writes with `--dry`.** It prints the HTTP requests without sending them. Show the preview to the user before deleting or bulk-changing data.
- **Never add `--force`** to skip confirmation prompts unless the user explicitly asked for it.
- **Keep the output small.** Use `--select`, `--filter` and `--pageSize` instead of fetching everything.
- Output is JSON when not run in a terminal. Use `--output csvheader` for tables, or `--view off` to disable views.
- Session modes limit what is allowed: `prod` = read only, `qual` = create and update, `dev` = create, update and delete.
  A "mode is disabled" error means the action is not allowed. Report it to the user and don't retry with other flags.

## Finding data

```sh
c8y devices list --name "sensor*" --type myType --select id,name,type
c8y devices list --query "name eq '*pump*' and has(c8y_IsDevice)" --pageSize 100
c8y devices list --lastMessageDateTo -1d --select id,name,c8y_Availability.status   # no data in the last day
c8y devices get --id 12345                    # by id
c8y devices get --id myDevice01               # devices can also be referenced by name
c8y inventory find --query "type eq 'c8y_DeviceGroup'"
c8y devices children list --id 12345 --childType device
c8y inventory count --fragmentType c8y_IsDevice # count instead of listing
```

- `--query` uses the Cumulocity inventory query language (server side).
- `--filter` filters on the client side after the response is received, e.g.
  `--filter "name like *pump*"`, `--filter "c8y_Firmware.version version >1.0.1"`, `--filter "has c8y_Position"`.
- `--select` picks properties and supports wildcards and aliases:
  `--select "id,name,serial:c8y_Hardware.serialNumber"`, `--select "id,c8y_Software*.**"`.
- Paging: `--pageSize` (max 2000) per page. `--includeAll` fetches every page (at 2000 per page), so combine it with `--select` and only use it when needed.
- Dates accept ISO 8601 or relative values: `-1h`, `-7d`, `0s` (now).

## Measurements, alarms, events and operations

```sh
c8y measurements list --device 12345 --dateFrom -1h --type c8y_Temperature
c8y measurements list --device 12345 --dateFrom -1d --valueFragmentType c8y_Temperature --pageSize 2000 --select time,c8y_Temperature.**
c8y alarms list --device 12345 --status ACTIVE --severity MAJOR
c8y alarms count --status ACTIVE
c8y events list --device 12345 --dateFrom -30min --type myEvent
c8y operations list --device 12345 --status PENDING
```

## Changing data (needs qual or dev mode)

```sh
c8y devices create --name test01 --type myType
c8y devices update --id 12345 --data "myFragment.value=1,name=new name"
c8y devices update --id 12345 --template "{c8y_SupportedOperations: ['c8y_Restart']}"
c8y measurements create --device 12345 --type myType --template "{c8y_Temperature:{T:{value: 25.0, unit:'C'}}}"
c8y operations create --device 12345 --data "c8y_Restart={}" --description "Restart device"
c8y operations wait --id 98765 --duration 60s
c8y alarms update --id 4567 --status CLEARED
c8y devices delete --id 12345                 # dev mode only
```

- `--data` accepts shorthand (`a.b=1,c=text`) or JSON. `--template` accepts jsonnet and is easier to quote.
- Setting a fragment to `null` removes it: `--data "myFragment=null"`.

## Running commands on a device (needs qual or dev mode)

When troubleshooting a device, run shell commands on it. With the MCP server, the user has to enable this
(`deviceShell.enabled` in `c8y_session`); if it is disabled, ask the user instead of trying other ways. Check what the device supports first, e.g.
`c8y inventory get --id 12345 --select c8y_SupportedOperations,c8y_RemoteAccessList`.

Via ssh, if Cumulocity remote access is configured for the device (uses the user's ssh keys or ssh agent):

```sh
c8y remoteaccess connect ssh --device 12345 -- systemctl status tedge-agent
c8y remoteaccess connect ssh --device 12345 --user root -- "sh -c 'journalctl -u tedge-agent -n 50 --no-pager'"
```

- Put the command after `--`. Quote commands which use shell syntax (pipes, `&&`, redirects).
- Only single commands are supported, not interactive sessions. Avoid commands which never finish, such
  as `journalctl -f` or `tail -f`; limit the output instead (`-n 50`).

Via a `c8y_Command` operation, if the device lists `c8y_Command` in `c8y_SupportedOperations`:

```sh
c8y operations create --device 12345 --description "Check disk space" --data "c8y_Command.text=df -h" \
  | c8y operations wait --status SUCCESSFUL --status FAILED --duration 2m --select status,c8y_Command.result,failureReason
```

- The output is in `c8y_Command.result`, and the reason for a failure in `failureReason`.

Both run on a real device, so say what a command does before running it, and prefer read-only commands
(status, logs, configuration) unless the user asks for a change.

## Pipelines

Commands accept the output of other c8y commands on stdin. The `id` (or `self`) of each piped object is used
as the main input of the next command, and each line is processed separately:

```sh
c8y devices list --name "sensor*" | c8y operations create --data "c8y_Restart={}" --description "Restart"
c8y devices list --type oldType --select id | c8y devices update --data "type=newType" --dry
c8y alarms list --device 12345 --status ACTIVE | c8y alarms update --status CLEARED
```

Always run a write pipeline with `--dry` first to confirm how many objects it affects.

## Generic REST requests

Use `c8y api` when no dedicated command exists. Paths are relative to the session's host:

```sh
c8y api /tenant/currentTenant
c8y api GET "/inventory/managedObjects?query=has(c8y_IsDevice)&pageSize=5"
c8y api POST /event/events --data "type=myType,text=hello,source.id=12345,time=2026-01-01T00:00:00Z"
```

## Troubleshooting

- `401` / `403`: the session is not logged in or lacks permissions. Ask the user; don't try other sessions.
- `mode is disabled`: the action needs a higher session mode. Report it; don't add `--sessionMode` or `--force`.
- `access to the file is denied`: the file is protected, e.g. a session file. Report it; don't try to read it another way.
- Unknown command or flag: check `c8y <resource> --help` for the available sub commands.
- Large or truncated output: add `--select`, `--filter` or a smaller `--pageSize`.
