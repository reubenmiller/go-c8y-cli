---
layout: default
category: Installation
order: 200
title: Creating bash aliases
---

## Install bash profile and dependencies

The following steps can be used to add custom c8y aliases to save in typing out the full command.

1. Install the bash completion package `bash-completion`

    **MacOS**

    ```sh
    brew install bash-completion@2
    ```

    **Debian**

    ```sh
    apt install bash-completion
    ```

    **Fedora**

    ```sh
    dnf install bash-completion
    ```

2. Download the c8y helper script

    ```sh
    curl -L https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/master/tools/c8y.profile.sh -o ~/c8y.profile.sh
    ```

3. Add the following line to your bach profile

    ```sh
    source ~/c8y.profile.sh
    ```

    Or view the command line

    ```sh
    echo "source ~/c8y.profile.sh" >> ~/.bashrc
    ```

4. Start a new bash console (or load your bash profile again)

    ```sh
    bash
    ```

    Try the 
    ```sh
    c8y <tab-tab>
    ```

## Creating custom bash aliases

1. Add a new alias definition to the `~/c8y.profile.sh`

    ```sh
    # create custom devices collection
    alias my_devices=c8y\ devices\ list --type "myCustomType"
    complete -F _complete_alias my_devices
    ```

2. Reload your bash session

3. Run your new custom alias

    ```sh
    my_devices
    ```
