
# Installing via a Linux Package Manager

## Ubuntu / Debian

1. Add the repository

    ```sh
    sudo sh -c "curl https://reubenmiller.jfrog.io/artifactory/api/security/keypair/public/repositories/c8y-debian | apt-key add -"
    sudo sh -c "echo 'deb https://reubenmiller.jfrog.io/artifactory/c8y-debian stable main' >> /etc/apt/sources.list"
    ```

    :::note
    The distribution is set to `stable` so that the same string can be used in all Debian based operating systems (i.e. Debian, Ubuntu etc.)
    :::

2. Update the repo and install go-c8y-cli

    ```sh
    sudo apt-get update
    sudo apt-get install go-c8y-cli
    ```

## Alpine Linux

1. Add the repository

    ```sh
    wget -O /etc/apk/keys/rmiller-rsa-signing.rsa.pub https://reubenmiller.jfrog.io/artifactory/api/security/keypair/public/repositories/c8y-alpine

    # Add the repo
    sh -c "echo 'https://reubenmiller.jfrog.io/artifactory/c8y-alpine/stable/main'" >> /etc/apk/repositories
    ```

2. Update the repo and install go-c8y-cli

    ```sh
    apk update
    apk add go-c8y-cli
    ```

## CentOS/RHEL/Fedora

1. Add a new repository using `vi`

    ```sh
    sudo vi /etc/yum.repos.d/artifactory.repo
    ```

    ```text
    [Artifactory]
    name=Artifactory
    baseurl=https://reubenmiller.jfrog.io/artifactory/c8y-rpm
    enabled=1
    gpgcheck=0
    gpgkey=https://reubenmiller.jfrog.io/artifactory/c8y-rpm/repodata/repomd.xml.key
    repo_gpgcheck=1
    ```

2. Update then install go-c8y-cli

    ```sh
    sudo yum update
    sudo yum install go-c8y-cli
    ```

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
