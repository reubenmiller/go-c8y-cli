---
category: Concepts - Extensions
title: Script based commands
---

An extension can include any number of commands. The structure of the commands is based on the folder structure, so you can group commands by placing them under the same sub folder. There is no limit to the number of sub folders, however you should keep it under 4-5 levels so it is not annoying for users to type.

Below shows some examples of commands provided by an extension called `c8y-myext` and how each command can be executed.

| Path | Command called via |
|-----------|------|
| `./commands/services/list` | `c8y myext services list` |
| `./commands/services/get` | `c8y myext services get` |
| `./commands/list` | `c8y myext list` |

The commands themselves can be written in any script-based language, e.g. `bash`, `python`, `ruby` etc., however they should include a [Shebang](https://en.wikipedia.org/wiki/Shebang_(Unix)) so that the correct shell interpreter can be called by the operating system.

Below is a simple example of a bash-based script.

```bash title="file: ./commands/list"
#!/usr/bin/env bash
set -e
echo "Hey look at me" >&2
```

:::note
Note: Make sure that your file is executable. On Unix, you can execute `chmod +x file_name` in the command line to make file_name executable. On Windows, you can run `git init -b main`, `git add file_name`, then `git update-index --chmod=+x file_name`.
:::

## General Tips

Below are some general tips to keep in mind when creating script based commands:

* Explicitly check for any dependencies (don't assume the user has these installed)
* Prefer bash over any other shell. Bash is installed by default on many different Operating Systems (though python might be a close second)
* Support the `--help|-h` flag so users can check how to use the command (don't rely on online documentation)
* Do not use `sudo` inside any scripts. If a script is called without the required elevated rights, then just log an error saying so 


## Bash Tips

The following sections detail tips when writing bash scripts. Some of the topics can also be applied to other scripting languages, however the examples are all shown in bash.

### Remove file extension

The script should not have a file extension as the name of the file is used in go-c8y-cli to list which commands are available to use.


:::danger bad

```bash
commands/myscript.sh
```

:::

:::tip good

```
commands/myscript
```

:::

### [Shebang](https://www.youtube.com/watch?v=dBUGfs9rwms) for better portability

A [shebang](https://en.wikipedia.org/wiki/Shebang_(Unix)) is the first line of a script which indicates which interpreters (e.g. shell) should be used to run the script.

It is recommended to use the `#!/usr/bin/env bash` shebang over `#!/bin/bash` as it allows users more control over which `bash` executable should be used if they have multiple bash interpreters installed.

```bash
#!/usr/bin/env bash
```

:::info
MacOS has bash version 3 installed by default. For reference bash 3.2 was released in 2006, and the more modern bash version 5.0 was released in 2019. By using the `#!/usr/bin/env bash` shebang, MacOS users can install a more recent version of bash using [homebrew](https://brew.sh/) and the newer version of bash will be used by default (providing the `PATH` variable points to the homebrew bin path).
:::


### Fail on unexpected errors

Use the `set -e` option to stop on unexpected errors. This is recommended because an unexpected errors can have some nasty side-effects as errors usually result in assigning an empty value to a variable, and downstream usage of the variable without validation could produce unexpected results (e.g. delete too many items etc.)

```bash
#!/usr/bin/env bash
set -e

echo "do something" >&2
```


### Print log messages to standard error

All log messages, user information and or progress indicators should be written to standard error (not standard output). This is because only standard output is piped by default, and generally piping log messages to any downstream executable does not make any sense as the messages are intended for users and not binaries.

You can use echo and the stream redirection syntax to write to standard error. Below shows an example of this.

```bash
echo "Running script-based command: $0" >&2
```

Alternatively you can create helper functions to log messages to standard error. The helper functions prepend the log level so that messages can be given different symantec meanings.

```bash
# Log helpers
warn()  { echo "WARN  $*" >&2; }
error() { echo "ERROR $*" >&2; }

# Use the helpers
warn "Something unexpected happened, but it's ok, we can still continue :)"
error "Oops, you did something wrong that we don't know how to fix"
```

```bash title="Output"
WARN  Something unexpected happened, but it's ok, we can still continue :)
ERROR Oops, you did something wrong that we don't know how to fix
```

### Use shellcheck to validate your script during development

[shellcheck](https://github.com/koalaman/shellcheck) is a great tool which checks for common mistakes/pitfalls. There are many IDE integrations available (e.g. VS Code extension etc.), so please use it, it will save you a world of pain, especially if you are not so experienced with writing bash scripts.

### Argument parsing

Support flag based argument parsing can be tricky in bash. You might be better off checking out some of the following projects to help you

|Name|Description|
|---|---|
|[docopt.sh](https://github.com/andsens/docopt.sh)|Automatically add bash parsing to a script from a doc string (at build time) |


However if you want to do parsing yourself, then you can use the following boiler plate:

```bash
#!/usr/bin/env bash
set -e

# Help text
help() {
  cat <<EOF
List items

Usage:
    c8y organizer devices list [FLAGS]

Examples:
    c8y organizer devices list --name "my name*" --onlyAgents

Flags:
    --name <string>       Match by name
    --onlyAgents          Only include agents   
EOF
}


NAME=${NAME:-}
AGENTS_ONLY=${AGENTS_ONLY:-"false"}
POSITIONAL_ARGS=()

#
# Parse Flags: --flag <value>, or boolean/switch flags: --help|-h
#
while [ $# -gt 0 ]; do
    case "$1" in
        # Flag which expects an argument, e.g. --name "value"
        --name)
            NAME="$2"
            shift
            ;;

        # Flag which does not expect an argument, e.g. --onlyAgents
        --onlyAgents)
            AGENTS_ONLY="true"
            ;;

        # Support showing the help when users provide '-h' or '--help'
        -h|--help)
            help
            exit 0
            ;;

        # Save positional arguments and restore them later on
        *)
            POSITIONAL_ARGS+=("$1")
            ;;
    esac
    shift
done

# Restore additional arguments which can then be referenced via "$@" and "$1" etc.
set -- "${POSITIONAL_ARGS[@]}"

echo "do something: NAME=${NAME}, AGENTS_ONLY=${AGENTS_ONLY}, OTHER_OPTIONS=$*"
```

### Scripts should be executable

A script based command must be made executable before go-c8y-cli can call it. The best place to do this is to make the script executable during development and commit the scripts to the git repository.

On MacOS/Linux you can make all scripts under the `commands/` folder executable by using this one-liner.

```bash
find commands/ -name "*" -exec chmod +x {} \;
```

Or if you are on Windows and are not using bash, then you can use the git command:

```bash
git update-index --chmod=+x commands/mycommand
```

Afterwards you will have to commit any changes to the file (yes changing the execution bit on files trigger a git change by default).


### Don't use sudo

Please don't use `sudo` inside any commands. If you need to run anything with elevated rights, log an error message indicating that the command needs to be run with elevated rights.

## Examples

### Example: List command with custom options

In this example a custom command is created to list the devices or agents in the platform. The user is given the option to search for devices or agents via using the `--onlyAgents` flag. The devices or agents can be filtered by name via the `--name <name>` flag.

A script based command is created by placing it under the `commands/` folder in the extension folder. You can create a command hierarchy by creating sub folder, for example `commands/devices/list` can be called via `c8y <ext_name> devices list`.

The script below shows the contents of the `list` command which is placed under the `commands/devices/` folder of the extension.

```bash title="file: commands/devices/list"
#!/usr/bin/env bash

# Stop on unexpected errors
set -e

# Logging helper functions
warn()  { echo "WARN  $*" >&2; }
error() { echo "ERROR $*" >&2; }

# Help text
help() {
  cat <<EOF
List items

Usage:
    c8y organizer devices list [FLAGS]

Examples:
    c8y organizer devices list --name "my name*" --onlyAgents

Flags:
    --name <string>       Match by name
    --onlyAgents          Only include agents   
EOF
}


NAME=${NAME:-}
AGENTS_ONLY=${AGENTS_ONLY:-"false"}
POSITIONAL_ARGS=()

#
# Parse Flags: --flag <value>, or boolean/switch flags: --help|-h
#
while [ $# -gt 0 ]; do
    case "$1" in
        # Flag which expects an argument, e.g. --name "value"
        --name)
            NAME="$2"
            shift
            ;;

        # Flag which does not expect an argument, e.g. --onlyAgents
        --onlyAgents)
            AGENTS_ONLY="true"
            ;;

        # Support showing the help when users provide '-h' or '--help'
        -h|--help)
            help
            exit 0
            ;;

        # Save positional arguments and restore them later on
        *)
            POSITIONAL_ARGS+=("$1")
            ;;
    esac
    shift
done

# Restore additional arguments which can then be referenced via "$@" and "$1" etc.
set -- "${POSITIONAL_ARGS[@]}"

## Validate arguments

# Check NAME property is not empty
if [ -z "$NAME" ]; then
  error "name should not be empty"
  help
  exit 1
fi

# Finally do what we came here to do! Pass any additional positional args using "$@" syntax
if [ "$AGENTS_ONLY" = "true" ]; then
  c8y agents list --name "$NAME" "$@"
else
  c8y devices list --name "$NAME" "$@"
fi
```

Assuming the command was placed under an extension called `organizer`, the command can be called via:

<CodeExample>

```sh
c8y organizer devices list --name "my device*"
```

</CodeExample>

The above script was built to pass extra flags/arguments provided by the user to the underlying `c8y devices|agents list` command, so we can provide the additional flags such as `--pageSize 100` will be passed to the other `c8y` commands.

<CodeExample>

```sh
c8y organizer devices list --name "my device*" --pageSize 100
```

</CodeExample>

:::note
go-c8y-cli has no way of knowing whether script based commands are support the extra flags or not, so tab completion will not be available.
:::
