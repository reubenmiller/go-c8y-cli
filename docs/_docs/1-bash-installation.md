---
layout: default
category: Installation - bash/zsh
order: 100
title: Installation
---

## Bash installation

To make the installation and future updates easier, a script is provided in the repository. The `c8y.profile.sh` script 

* `set-session` to help switch between c8y sessions and set the appropriate environment variables
* Bash completion
* Bash completion for aliases using [complete_alias](https://raw.githubusercontent.com/cykerway/complete-alias/master/complete_alias)
* Install/Update c8y to latest version
* Adds c8y aliases such as `devices`, `mo`, `ops`

### Installing

1. Download the c8y profile helper script

    ```sh
    curl -L https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/develop/tools/bash/c8y.profile.sh -o ~/c8y.profile.sh

    # make the script executable
    chmod +x ~/c8y.profile.sh
    ```

2. Add the following line to your bash profile `~/.bashrc`

    ```sh
    source ~/c8y.profile.sh
    ```

3. Reload your bash profile

    ```sh
    source ~/.bashrc
    ```

4. Install the latest c8y version

    ```sh
    c8y-update
    ```

    By default the script will add the binary to your home bin folder `~/bin`.

    Alternatively the version and installation path can be controlled via arguments to `c8y-update`. Below shows the example usage.

    **Install latest version in specific folder**

    ```sh
    c8y-update latest ./
    ```

    **Install a specific version**

    ```sh
    c8y-update v1.4.1
    ```

    The list of versions is available from the [github release page](https://github.com/reubenmiller/go-c8y-cli/releases)

---

## Zsh (oh-my-zsh) installation

Similar to `c8y.profile.sh` for bash, there is a pre-configured zsh plugin `c8y.plugin.zsh` which provides the same functions but for zsh.

### Installing

1. Create the custom plugin folder

    ```sh
    mkdir -p ~/.oh-my-zsh/custom/plugins/c8y
    ```

2. Download the c8y zsh plugin

    ```sh
    curl -L https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/master/tools/bash/c8y.plugin.zsh -o ~/.oh-my-zsh/custom/plugins/c8y/c8y.plugin.zsh

    ```

3. Add the `c8y` to your plugin list in `~/.zshrc`

    For example

    ```sh
    plugins=(git c8y)
    ```

4. Reload your zsh profile to load the new plugin

    ```sh
    source ~/.zshrc
    ```

5. Install/update the latest c8y version

    ```sh
    c8y-update
    ```

6. Reload your profile again to enable the completions

    ```sh
    source ~/.zshrc
    ```

7. Now c8y is ready to use. Future updates can be done using

    ```sh
    c8y-update

    # reload to update any auto completions
    source ~/.zshrc
    ```

---

## Manual installation

Since the c8y cli tool is a golang binary, it is very portable and can therefore be downloaded manually and added to your `PATH` environment variable.

However if you don't install the `c8y.profile.sh` script or the zsh `c8y` plugin, then you are reponsible for repeating this procedure when updating c8y, and when setting/switching c8y sessions.

1. Download the latest version

    ```sh
    curl -L https://github.com/reubenmiller/go-c8y-cli/releases/download/latest/c8y.windows.exe -o c8y
    ```

    **Note**

    You will need to change the `1.1.0` version if a newer version is released. You can browser the available versions by looking at the [PSc8y release page](https://www.powershellgallery.com/packages/PSc8y), then just change the `1.1.0` to one of the listed versions.

    The go c8y cli tool is used internally by the PowerShell PSc8y module, that is why it can also be downloaded from the PowerShell Gallery website (even though you don't have PowerShell)

1. Extract the c8y binary (relevant to your OS) from within downloaded zip file

    **Linux**

    ```sh
    curl -L https://github.com/reubenmiller/go-c8y-cli/releases/download/latest/c8y.linux -o ./c8y
    ```

    **MacOS**

    ```sh
    curl -L https://github.com/reubenmiller/go-c8y-cli/releases/download/latest/c8y.macos -o ./c8y
    ```

    **Windows**

    ```sh
    curl -L https://github.com/reubenmiller/go-c8y-cli/releases/download/latest/c8y.windows.exe -o ./c8y
    ```


1. Copy the file to a path inside your `PATH` variable

    ```sh
    chmod +x ./c8y
    sudo cp ./c8y /usr/local/bin/
    ```

    Or you can add the current path (where the c8y binary is) to your `PATH` variable

    ```sh
    chmod +x ./c8y
    export PATH=$(pwd):$PATH
    ```

1. Check if the c8y binary is now callable from anywhere by checking the version

    ```sh
    c8y version
    ```

    **Response**

    ```plaintext
    Cumulocity command line tool
    v1.4.1 -- HEAD
    ```

1. Add completions to your profile:

    **Bash**

    ```sh
    echo "source <(c8y completion bash)" >> ~/.bashrc
    source ~/.bashrc
    ```

    **Zsh (oh-my-zsh)**

    ```sh

    mkdir -p ~/.oh-my-zsh/completions
    c8y completion zsh > ~/.oh-my-zsh/completions/_c8y
    source ~/.zshrc
    ```

    **Note**

    Try out the completions by entering

    ```sh
    c8y ver<tab>
    ```

1. Setting / switching c8y sessions

    If you don't have a session then see the [Getting started](https://reubenmiller.github.io/go-c8y-cli/docs/2-getting-started-bash/) section.

    ```sh
    export C8Y_SESSION=$( c8y sessions list )
    ```

    **Note**

    It is highly recommended that you use the `c8y.profile.sh` or `c8y` zsh plugin, as these include helpers to set the c8y session which improve your c8y experience.

---

## Other recommended tools

Since the output of the c8y cli tool is mainly json, it is highly recommended that you install the json cli tool `jq` to help formatting the output.

**Example: Get the id of each devices from a query**

```sh
c8y devices list | jq -r ".[].id"
```

See the [jq website](https://stedolan.github.io/jq/download/) for details how to install it on your operating system.
