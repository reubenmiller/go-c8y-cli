---
title: Shell
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

:::caution
The install script currently requires you to have `jq` installed. `jq` is a cli json parsing tool which is recommended to have anyways to do complex json manipulations that you might need during your daily use of `c8y`.

See the [jq website](https://stedolan.github.io/jq/download/) for details how to install it on your operating system.
:::

## Installation

1. Clone the addons repository containing the install script, views and some example templates.

    ```sh
    git clone https://github.com/reubenmiller/go-c8y-cli-addons.git ~/.go-c8y-cli
    ```

    :::info
    This repository is not strictly required to run `c8y` however it makes is much more useful as it provides some  default views and templates to get the most out of `c8y`. The defaults also show you how you can create your own custom views.
    :::

2. Install go-c8y-cli binary

    **Option 1: without sudo**

    ```sh
    ~/.go-c8y-cli/install.sh
    ```

    :::info
    You will need to add the installation folder to your `PATH` environment variable, however the install script should print out a message with an example of what you need to do.

    But just in case, you can will need to add the `bin` folder to your shell profile. For bash, the default profile is called `.bashrc`

    ```bash title="bash profile: ~/.bashrc"
    export PATH="/home/myuser/bin:$PATH"
    ```

    Reload your shell, or source your profile directly. `source ~/.bashrc`
    :::

    **Option 2: with sudo**
    If you want to make the c8y globally available, then you can run it using sudo, which will install it under `/usr/local/bin`.

    ```sh
    sudo -E ~/.go-c8y-cli/install.sh
    ```

    :::tip
    If you get an error that the `install.sh` script is not executable, then run the following and try again.

    ```sh
    chmod +x ~/.go-c8y-cli/install.sh
    ```
    :::

3. Verify that the `c8y` binary is executable and can be found on the command line 

    ```bash
    which c8y
    ```

    ```text title="Output"
    # If installed without sudo
    /home/myuser/bin/c8y
    
    # If installed with sudo
    /usr/local/bin/c8y
    ```

    :::tip
    Try closing your console and re-opening it so you can be sure that your setup will work next time
    :::

3. Now go to the [Getting started](../gettingstarted) section for instructions how to use it

---

## Other recommended tools

### jq

Since the output of the c8y cli tool is mainly json, it is highly recommended that you install the json cli tool `jq` to help formatting the output.

#### Example: Get the id of each devices from a query

```bash
c8y devices list --select id

# Or csv output using
c8y devices list --select id --output csvheader
```

If you are more familiar with the popular `jq` tool, then you can use it to extract information that you need.

```bash
c8y devices list | jq -r ".id"
```
