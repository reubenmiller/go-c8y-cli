---
layout: default
category: Getting started
title: Shell
---

Before getting started, you need to configure the Cumulocity session which has the details which Cumulocity platform and authentication should be used for each of the commands/requests. This operation only needs to be done once.

1. Create a new session:

    ```bash
    c8y sessions create \
        --host "https://mytenant.eu-latest.cumulocity.com" \
        --username "myUser@me.com" \
        --type dev
    ```

    You will be prompted for your password. Alternatively you can also enter the password using the `--password` argument.

    You may also provide a more meaningful session name by using the `--name` argument.

    The `--type` parameter indicates what kind of tenant you are using which directly controls which commands are enabled by default, and which are disabled. `dev` will enable all commands, whereas `prod` only allows GET commands.

    Alternatively you can create a session by creating a json file in your `~/.cumulocity` folder.

    **Example**

    **~/.cumulocity/session1.json**

    ```json
    {
        "host": "example01.cumulocity.eu-latest.com",
        "username": "hello.user@example.com",
        "password": "mys3cureP4assw!rd",
    }
    ```

2. Activate the session using the interactive session selector:

    ```bash
    set-session
    ```

    The list of sessions can be filtered by adding additional filter terms when calling `set-session`. If only 1 session is found, then it will be automatically selected without the user having to confirm the selection.

    ```bash
    set-session myname partial_string
    ```

3. Test your credentials by getting your current user information from the platform:

    ```bash
    c8y currentuser get

    # or list devices
    c8y devices list
    ```

    **Note**

    If your credentials are incorrect, then you can update the session file stored in the `~/.cumulocity` directory

4. Now you're ready to go. You can get a list of available commands by using the help menu:

    ```bash
    c8y help
    ```

## Switching sessions

The sessions can be changed again by using the interactive session selector

```bash
set-session
```

Alternatively you can switch sessions by calling the c8y binary directly (without the set-session helper)

```bash
# Set session interactively
eval $( c8y sessions set --shell=auto )

# Set a session to an already known json path.
eval $( c8y sessions set --shell=auto --session=/my/path/session.json )
```

#### Switching session for a single command

A single command can be redirected to use another session by using the `--session <name>` parameter. A full file path can be provided or just the file name for a file located in session directory defined by the `C8Y_SESSION_HOME` environment variable.

```bash
c8y devices list --session myothersession.json
```
