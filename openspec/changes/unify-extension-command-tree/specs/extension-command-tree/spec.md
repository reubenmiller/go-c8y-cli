## ADDED Requirements

### Requirement: Extension commands are namespaced under the extension name
The system SHALL register every command provided by an installed extension under `c8y <extension-name>`, where `<extension-name>` is the extension's name without the `c8y-` prefix. An extension with no commands SHALL NOT add a root command.

#### Scenario: Extension with commands
- **WHEN** extension `c8y-analytics` provides at least one command
- **THEN** `c8y analytics` is listed in `c8y --help`

#### Scenario: Extension with no commands
- **WHEN** extension `c8y-viewsonly` provides only views and templates
- **THEN** no `c8y viewsonly` command is registered

### Requirement: Command groups from different sources are merged
The system SHALL merge command groups that have the same path within an extension into a single group, regardless of whether the group was declared by an API spec file or implied by a folder under `commands/`.

#### Scenario: API group and script folder with the same name (#585)
- **WHEN** an extension has an API spec with `group.name: management` that declares command `health`, and a script at `commands/management/logs`
- **THEN** both `c8y <ext> management health` and `c8y <ext> management logs` run their respective commands
- **AND** `c8y <ext> --help` lists exactly one `management` group

#### Scenario: Two API spec files with the same group name
- **WHEN** `api/a.yaml` and `api/b.yaml` both declare `group.name: devices`, with commands `list` and `get` respectively
- **THEN** `c8y <ext> devices list` and `c8y <ext> devices get` both work
- **AND** only one `devices` group is registered

### Requirement: API specs can define nested command groups
The system SHALL split an API spec's `group.name` on whitespace and register its commands under one nested group per word. A spec with no commands SHALL still declare its group, so it can provide the description of a parent group.

#### Scenario: Nested group from a group name with spaces (#562)
- **WHEN** an API spec declares `group.name: inventory items` with commands `list` and `get`
- **THEN** `c8y <ext> inventory items list` and `c8y <ext> inventory items get` run those commands
- **AND** `c8y <ext> inventory --help` lists `items`

#### Scenario: Parent group description from a spec without commands
- **WHEN** one API spec declares `group.name: inventory items` with commands, and another declares `group.name: inventory` with `description: Manage inventory` and no commands
- **THEN** `c8y <ext> --help` shows `inventory` with the description `Manage inventory`

#### Scenario: Group without any commands
- **WHEN** the only spec for group `inventory` has no commands and no nested groups contain commands
- **THEN** no `inventory` group is registered

### Requirement: Script commands are registered at any folder depth
The system SHALL register an executable file at `commands/<p1>/<p2>/.../<name>` as `c8y <ext> <p1> <p2> ... <name>`, creating every intermediate group.

#### Scenario: Top-level script
- **WHEN** an extension has `commands/hello`
- **THEN** `c8y <ext> hello` runs that file

#### Scenario: Deeply nested script
- **WHEN** an extension has `commands/a/b/c/run`
- **THEN** `c8y <ext> a b c run` runs that file
- **AND** `c8y <ext> a --help`, `c8y <ext> a b --help`, and `c8y <ext> a b c --help` each list the next level

#### Scenario: Hidden files are ignored
- **WHEN** an extension has `commands/.gitkeep`
- **THEN** no command is registered for it

### Requirement: Leaf command conflicts are resolved by fixed precedence
When more than one source defines a runnable command at the same path within an extension, the system SHALL keep the command from the highest-precedence source, in the order API spec, then script, and SHALL log an info-level message (shown with `--verbose`) that names the kept and dropped origins, relative to the extension folder.

#### Scenario: API command and script with the same path
- **WHEN** an API spec declares `management health` and a script exists at `commands/management/health`
- **THEN** `c8y <ext> management health` runs the API command
- **AND** running with `--verbose` logs a message that includes both the spec file path and the script path
- **AND** no message is shown without `--verbose`

### Requirement: Groups take precedence over leaves at the same path
When one source defines a runnable command at a path and another source defines a group at the same path, the system SHALL keep the group and its children, drop the runnable command, and log an info-level message.

#### Scenario: Script file shadowed by a group
- **WHEN** a script exists at `commands/management` and an API spec declares `group.name: management`
- **THEN** `c8y <ext> management` is a group containing the API commands
- **AND** running with `--verbose` logs a message naming the dropped script

### Requirement: Group descriptions prefer explicit text
The system SHALL use the description from an API spec `group` for a merged group. A group that only comes from a `commands/` folder SHALL use a generated description.

#### Scenario: Merged group description
- **WHEN** an API spec declares `group.name: management` with `description: Manage analytics`, and a script folder `commands/management/` also exists
- **THEN** `c8y <ext> --help` shows `management` with the description `Manage analytics`

### Requirement: Command registration is deterministic
The system SHALL produce the same command tree, including conflict outcomes, regardless of file system enumeration order.

#### Scenario: Repeated startup
- **WHEN** c8y starts twice with the same installed extensions
- **THEN** the command tree and any conflict messages are identical both times
