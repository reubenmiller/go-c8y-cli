
# Installing via a Linux Package Manager

## Ubuntu / Debian

1. Install the apt dependencies

    ```
    sudo apt-get install -y curl gnupg2 apt-transport-https
    ```

    :::note
    The command requires you to already have `sudo` installed. Most installations will have sudo by default so you don't need to do anything, however if you don't have it, install it via apt when running as a root user.
    :::

2. Configure the repository

    **Debian >=9 and Ubuntu >= 16.04**

    ```bash
    curl https://reubenmiller.github.io/go-c8y-cli-repo/debian/PUBLIC.KEY | gpg --dearmor | sudo tee /usr/share/keyrings/go-c8y-cli-archive-keyring.gpg >/dev/null
    sudo sh -c "echo 'deb [signed-by=/usr/share/keyrings/go-c8y-cli-archive-keyring.gpg] http://reubenmiller.github.io/go-c8y-cli-repo/debian stable main' >> /etc/apt/sources.list"
    ```

    :::note
    This step does not make use of `apt-key` as it has been deprecated. The gpg key is stored in an individual store only related to the go-c8y-cli repository, and it is linked via the apt.source settings using the `signed-by` property.
    :::

    **Debian <=8 and Ubuntu <= 14.04**

    ```bash
    curl https://reubenmiller.github.io/go-c8y-cli-repo/debian/PUBLIC.KEY | sudo apt-key add -
    sudo sh -c "echo 'deb https://reubenmiller.github.io/go-c8y-cli-repo/debian stable main' >> /etc/apt/sources.list"
    ```

3. Update the repo then install/update `go-c8y-cli`

    ```bash
    sudo apt-get update
    sudo apt-get install go-c8y-cli
    ```

## Alpine Linux

**Note**

The following commands require sudo. If you don't have `sudo` installed, then remove the `sudo` from the command, and run as root user.

1. Install the apk dependencies

    ```bash
    sudo apk add wget
    ```

2. Configure the repository

    ```bash
    sudo wget -O /etc/apk/keys/reuben.d.miller\@gmail.com-61e3680b.rsa.pub https://reubenmiller.github.io/go-c8y-cli-repo/alpine/PUBLIC.KEY

    # Add the repo
    sudo sh -c "echo 'https://reubenmiller.github.io/go-c8y-cli-repo/alpine/stable/main'" >> /etc/apk/repositories
    ```

3. Update the repo then install/update `go-c8y-cli`

    ```bash
    sudo apk update
    sudo apk add go-c8y-cli
    ```

4. Follow the instructions to [setup your shell profile](/docs/installation/shell-installation#setting-up-your-shell-profile)


## CentOS/RHEL/Fedora

1. Configure the repository 

    Create a new file using your editor of choice (vi, vim, nano etc.)

    ```sh
    sudo vi /etc/yum.repos.d/go-c8y-cli.repo
    ```

    Then add the following contents to it and save the file.

    ```text title="/etc/yum.repos.d/go-c8y-cli.repo"
    [go-c8y-cli]
    name=go-c8y-cli packages
    baseurl=https://reubenmiller.github.io/go-c8y-cli-repo/rpm/stable
    enabled=1
    gpgcheck=1
    gpgkey=https://reubenmiller.github.io/go-c8y-cli-repo/rpm/PUBLIC.KEY
    ```

2. Update the repo then install/update `go-c8y-cli`

    ```sh
    sudo dnf update
    sudo dnf install go-c8y-cli
    ```

3. Follow the instructions to [setup your shell profile](/docs/installation/shell-installation#setting-up-your-shell-profile)

**Note**

You can install `go-c8y-cli` via `yum` by just replacing `dnf` with `yum` in the above commands.


## Add plugin to your profile

To enable command completion, the following needs to be added to your shell. Follow the instructions for the shell that you want to work with.

```sh
# bash
echo 'source "/etc/go-c8y-cli/shell/c8y.plugin.sh"' >> ~/.bashrc

# zsh
echo 'source "/etc/go-c8y-cli/shell/c8y.plugin.zsh"' >> ~/.zshrc

# fish
echo 'source "/etc/go-c8y-cli/shell/c8y.plugin.fish"' >> ~/.config/fish/config.fish
```

:::info
Launch your shell again for it to take effect
:::
