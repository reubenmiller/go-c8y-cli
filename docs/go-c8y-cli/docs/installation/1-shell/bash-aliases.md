---
layout: default
category: Installation - shell
order: 200
title: Creating aliases
---

### Shell

#### Install bash profile and dependencies

The following steps can be used to add custom c8y aliases to save in typing out the full command.

1. Install the bash completion package `bash-completion`

    **MacOS**

    ```bash
    brew install bash-completion@2
    ```

    **Debian**

    ```bash
    apt install bash-completion
    ```

    **Fedora**

    ```bash
    dnf install bash-completion
    ```

2. Download the c8y helper script

    ```bash
    curl -L -o ~/c8y.plugin.sh \
    https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/master/tools/shell/c8y.plugin.sh
    ```

3. Add the following line to your bash profile `~/.bashrc`

    ```bash
    source ~/c8y.plugin.sh
    ```

    Or using the command line

    ```bash
    echo "source ~/c8y.plugin.sh" >> ~/.bashrc
    ```

4. Reload your bash profile again

    ```bash
    source ~/.bashrc
    ```

    Test that the completion works with c8y

    ```bash
    c8y <tab-tab>
    ```

    Then, check if it works with aliases

    ```bash
    devices --<tab-tab>
    ```

#### Creating custom bash aliases

1. Add a new alias definition to the `~/c8y.plugin.sh`

    ```bash
    # create custom devices collection
    alias my_devices=c8y\ devices\ list --type "myCustomType"
    complete -F _complete_alias my_devices
    ```

2. Reload your bash session

3. Run your new custom alias

    ```bash
    my_devices
    ```

---

### Zsh

#### Create a custom operation to get all FAILED operations

1. Open your zsh profile `~/.zshrc` and add a custom alias

    ```bash
    alias failedops="c8y operations list --status FAILED"
    ```

2. Save the changes to your zsh profile

3. Reload the profile

    ```bash
    source ~/.zshrc
    ```

4. Try out your new alias

    ```bash
    failedops
    ```

    ```bash
    failedops --agent "test*"
    ```
