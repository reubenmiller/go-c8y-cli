# MCP server: choose the session mode in the picker

Status: proposal, prototype on `feat/mcp-session-mode`

## Problem

`c8y mcp serve` gives the agent the lower of `--mode` and the session's own mode
(`settings.session.mode`, prod if unset). The session picker only shows that mode.
To give the agent a different mode, the user has to edit the session file or restart
the server with other flags. A write blocked by the mode is rejected outright, with no
way to allow it just once.

Users, and power users in particular, want to:

- pick the mode together with the session in the picker,
- change the mode of the current session quickly, without logging in again,
- allow a single command that needs a higher mode.

## Model: ceiling, default, override

| Level | Source | Who can change it |
|---|---|---|
| Ceiling | `--mode` or `mode:` in the policy file. Default: `dev` | Whoever starts the server. Never exceeded |
| Default | `settings.session.mode` of the session (prod if unset) | The session file |
| Override | Picker, `mode` prompt, "switch for this session" in an approval | The user only, never the agent |

```
effective = min(ceiling, override if set, else session mode)
```

- CI mode still counts as dev.
- Overrides are kept in memory per session file for the lifetime of the server, so
  switching back to a session keeps its override. They are never written to the
  session file.
- `--mode` keeps its meaning as a hard maximum. New: when `--mode` is unset, the user
  can raise the mode above the session's own mode. Start the server with `--mode prod`
  (or the session's mode) to keep today's behavior.

Changing the mode only needs a copy of the `Runner` with another `Mode` and the same
`Auth`, so it never logs in again.

## 1. Picker

- Each session has a segmented control `prod · qual · dev`, preset to the session's own
  mode (or its current override). Modes above the ceiling are disabled. Modes above the
  session's own mode are marked as an override.
- The current session offers "Change mode only", which keeps the logged-in session.
- Keyboard: `↑`/`↓` move, `←`/`→` lower or raise the mode (no wrap-around, so one key
  press never goes from prod to dev), `1`/`2`/`3` set the mode, `Enter` selects,
  typing filters by name.
- The shortcuts are shown on the page: a hint line under the sessions (only with
  JavaScript, as the shortcuts need it), and a help dialog on `?` which also lists the
  `mode` and `switch_session` prompts. The controls announce their keys to screen
  readers (`aria-keyshortcuts`).

## 2. Prompts with arguments

MCP prompts are started by the user, not the agent.

- `mode <prod|qual|dev|reset>`: changes the mode of the current session at once, with
  no browser. `reset` returns to the session's own mode.
- `switch_session [session] [mode]`: with a session, switches straight away if it
  needs no passphrase or code, otherwise opens the picker at its unlock step. Without
  arguments it opens the picker as before.
- Arguments are completed (`completion/complete`): session names, and the modes
  allowed under the ceiling.
- The prompt result tells the agent what changed, e.g. "Mode is now dev (override; the
  session's own mode is prod)".

## 3. Allow once

When a write needs a higher mode than the effective one (POST/PUT: qual, DELETE: dev):

- With elicitation, the existing approval prompt is extended, not duplicated:
  "This needs dev; the session is in qual", with **Allow once**, **Switch to dev for
  this session**, or decline. The signed, single-use approval key includes the
  required mode. Allow once runs that one command with a runner copy at the higher
  mode, still capped by the ceiling.
- Without elicitation, or with approval `client` or `none`: no escalation. The error
  asks the agent to have the user run the `mode` prompt.
- The required mode comes from the dry run preview (HTTP methods), falling back to
  the command verb (`delete*` means dev).
- Approval is still asked in every mode. An override never skips approval.

## Visibility

- `c8y_session` adds `modeSource` (`session`, `override`), `sessionMode` and
  `maxMode`.
- The server log records every change, e.g.
  `mode override: prod -> dev (session X, via picker)`.

## Enforcement in one place

Mode checks, required mode classification and the choice of runner move from the
tool handlers into one shared layer used by every entry point. The read and write
tools use it today; code mode (below) will use it too.

## Code mode (later)

Planned: the agent submits a script which calls many commands in one tool call.

- The session and mode are fixed when a run starts. Changes apply to the next run.
- "Allow once" becomes "allow this run": a write above the run's mode fails inside the
  script. Optionally, the agent declares the mode it needs and the user approves the
  script and mode together, bound to the script's hash.
- Consider a per-run cap on deletes, or a qual ceiling for code mode unless the user
  opts in to dev.
- An expiring override (`modeOverrideTTL`) becomes more useful and could be on by
  default above the session's own mode.

## Decisions taken for the prototype

1. Raising above the session's own mode is allowed by default, capped by `--mode`.
2. Overrides persist per session until the server stops.
3. No expiry yet (`modeOverrideTTL` left for code mode).
