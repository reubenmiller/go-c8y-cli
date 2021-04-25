---
title: Shell
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

1. Clone the addons directory containing the install script, views and some example templates.

    ```sh
    git clone https://github.com/reubenmiller/go-c8y-cli-addons.git ~/.go-c8y-cli
    cd ~/.go-c8y-cli
    ```

    :::info
    This repository is not strictly required to run `c8y` however it makes is much more useful as it helps you should how views and templates can be used so you can get the most out of `c8y`
    :::

2. Install go-c8y-cli binary

    **Option 1: without sudo**

    ```sh
    ~/.go-c8y-cli/install.sh
    ```

    **Option 2: with sudo**
    If you want to make the c8y globally available, then you can run it using sudo, which will install it under `/usr/local/bin`.

    ```sh
    sudo -E ~/.go-c8y-cli/install.sh
    ```

3. Create a new session (if you do not already have one)
    
    ```sh
    c8y sessions create --host myinfo --username johnsmith@example.com --type dev
    ```

4. Activate the newly created profile

    ```sh
    set-session
    ```

5. Check the profile by showing your current user information

    ```sh
    c8y currentuser get
    ```

    ```json title="output"
    | id                          | email                       | passwordstrength | lastPasswordChange            | shouldResetPassword |
    |-----------------------------|-----------------------------|------------------|-------------------------------|---------------------|
    | johnsmith@example.com       | johnsmith@example.com       |                  | 2020-12-16T16:26:34.803Z      | false               |
    ```

---

## Other recommended tools

Since the output of the c8y cli tool is mainly json, it is highly recommended that you install the json cli tool `jq` to help formatting the output.

**Example: Get the id of each devices from a query**

```bash
c8y devices list --select id

# Or csv output using
c8y devices list --select id --output csvheader
```

If you are more familiar with the popular `jq` tool, then you can use it to extract information that you need.

```bash
c8y devices list | jq -r ".id"
```
See the [jq website](https://stedolan.github.io/jq/download/) for details how to install it on your operating system.
