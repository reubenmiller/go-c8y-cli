---
layout: default
category: Installation
order: 100
title: Bash
---

1. Download the latest version

    ```sh
    curl -L https://www.powershellgallery.com/api/v2/package/PSc8y/1.1.0 -o c8y.zip
    ```

    **Note**

    You will need to change the "1.1.0" version if a newer version is released. You can browser the available versions by looking at the [PSc8y release page](https://www.powershellgallery.com/packages/PSc8y), then just change the "1.1.0" to one of the listed versions.

    The go c8y cli tool is used internally by the PowerShell PSc8y module, that is why it can also be downloaded from the PowerShell Gallery website (even though you don't have PowerShell)

1. Extract the c8y binary (relevant to your OS) from within downloaded zip file

    **Linux**

    ```sh
    unzip -p c8y.zip Dependencies/c8y.linux > ./c8y
    ```

    **MacOS**

    ```sh
    unzip -p c8y.zip Dependencies/c8y.macos > ./c8y
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
    v0.7.0-345-g164bcec -- master
    ```

1. Add completions to your profile:

    **Bash**

    ```sh
    echo "source <(c8y completion bash)" >> ~/.bashrc
    source ~/.bashrc
    ```

    **Zsh**

    ```sh
    echo "source <(c8y completion zsh)" >> ~/.zshrc
    echo "complete -F _c8y c8y" >> ~/.zshrc

    source ~/.zshrc
    ```

    **Note**

    Try out the completions by entering

    ```sh
    c8y ver<tab>
    ```
