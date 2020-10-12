---
layout: default
category: Getting started
title: Bash
---

Before getting started, you need to configure the Cumulocity session which has the details which Cumulocity platform and authentication should be used for each of the commands/requests. This operation only needs to be done once.

1. Create a new session:

    ```sh
    c8y sessions create \
        --host "https://mytenant.eu-latest.cumulocity.com" \
        --tenant "myTenant" \
        --username "myUser@me.com"
    ```

    You will be prompted for your password. Alternatively you can also enter the password using the `--password` argument.

    You may also provide a more meaningful session name by using the `--name` argument.

    Alternatively you can create a session by creating a json file in your `~/.cumulocity` folder.

    **Example**

    **~/.cumulocity/session1.json**

    ```json
    {
        "host": "https://example01.cumulocity.eu-latest.com",
        "tenant": "t12345",
        "username": "hello.user@example.com",
        "password": "mys3cureP4assw!rd",
        "description": "",
        "useTenantPrefix": true,
        "microserviceAliases": {}
    }
    ```

2. Activate the session using the interactive session selector:

    ```sh
    set-session
    ```

    Or you can manually set the path to the session (json) file by setting the `C8Y_SESSION` environment variable.

    ```sh
    export C8Y_SESSION=/path/to/any/session/session1.json
    ```

3. Test your credentials by getting your current user information from the platform:

    ```sh
    c8y users getCurrentUser
    ```

    **Note**

    If your credentials are incorrect, then you can update the session file stored in the `~/.cumulocity` directory

4. Now you're ready to go. You can get a list of available commands by using the help menu:

    ```sh
    c8y help
    ```

## Switching sessions

The sessions can be changed again by using the interactive session selector

```sh
set-session
```

Alternatively you can switch sessions manually via setting the `C8Y_SESSION` environment variable.

```sh
# Set session interactively
export C8Y_SESSION=$( c8y sessions list )

# Set a session to an already known json path.
export C8Y_SESSION=/my/path/session.json
```
