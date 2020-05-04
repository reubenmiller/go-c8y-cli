---
layout: default
category: Getting started
title: Bash
---

# Getting started

Before getting started, you need to configure the Cumulocity session which has the details which Cumulocity platform and authentication should be used for each of the commands/requests. This operation only needs to be done once.

1. Create a new session:

    ```sh
    c8y sessions create \
        --host "http://mytenant.eu-latest.cumulocity.com" \
        --tenant "myTenant" \
        --username "myUser@me.com"
    ```

    You will be prompted for your password. Alternatively you can also enter the password using the `--password` argument.

    You may also provide a more meaningful session name by using the `--name` argument.

2. Activate the session using the interactive session selector:

    ```sh
    # bash
    export C8Y_SESSION=$( c8y sessions list )
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
export C8Y_SESSION=$( c8y sessions list )
```
