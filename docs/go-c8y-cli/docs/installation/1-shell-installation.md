---
title: Shell
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Installation

**go-c8y-cli** is available as a pre-built binary which can be installed and updated via a package manager. Please following the instructions for your operating system.

It is recommended to install **go-c8y-cli** using a package manager as it makes it easier to update it in the future, and it will be available for all users.

### Debian / Ubuntu (apt)

1. Install the apt dependencies

    ```
    sudo apt-get install -y curl gnupg2 apt-transport-https
    ```

    :::note
    The command requires you to already have `sudo` installed. Most installations will have sudo by default so you don't need to do anything, however if you don't have it, install it via apt when running as a root user.
    :::

2. Configure the repository

    <Tabs
    groupId="debian-os"
    defaultValue="debian"
    values={[
        { label: 'Debian >=9 and Ubuntu >= 16.04', value: 'debian', },
        { label: 'Debian <=8 and Ubuntu <= 14.04', value: 'debian_legacy', }
    ]
    }>
    <TabItem value="debian">

    ```bash
    curl https://reubenmiller.github.io/go-c8y-cli-repo/debian/PUBLIC.KEY | gpg --dearmor | sudo tee /usr/share/keyrings/go-c8y-cli-archive-keyring.gpg >/dev/null
    sudo sh -c "echo 'deb [signed-by=/usr/share/keyrings/go-c8y-cli-archive-keyring.gpg] http://reubenmiller.github.io/go-c8y-cli-repo/debian stable main' >> /etc/apt/sources.list"
    ```

    </TabItem>
    <TabItem value="debian_legacy">

    ```bash
    curl https://reubenmiller.github.io/go-c8y-cli-repo/debian/PUBLIC.KEY | sudo apt-key add -
    sudo sh -c "echo 'deb https://reubenmiller.github.io/go-c8y-cli-repo/debian stable main' >> /etc/apt/sources.list"
    ```

    </TabItem>
    </Tabs>

3. Update the repo then install/update **go-c8y-cli**

    ```bash
    sudo apt-get update
    sudo apt-get install go-c8y-cli
    ```

4. Install go-c8y-cli helper functions and reload your shell

    ```bash
    c8y cli install
    ```

5. Go to the [Getting started section](/docs/gettingstarted/) for the next steps

---

### CentOS/RHEL/Fedora (dnf/yum)

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

2. Update the repo then install/update **go-c8y-cli**

    ```sh
    sudo dnf update
    sudo dnf install go-c8y-cli
    ```

3. Install go-c8y-cli helper functions and reload your shell

    ```bash
    c8y cli install
    ```

4. Go to the [Getting started section](/docs/gettingstarted/) for the next steps

:::note
You can install **go-c8y-cli** via `yum` by just replacing `dnf` with `yum` in the above commands.
:::

---

### Alpine (apk)

:::note
The following commands require sudo. If you don't have `sudo` installed, then remove the `sudo` from the command, and run as root user.
:::

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

3. Update the repo then install/update **go-c8y-cli**

    ```bash
    sudo apk update
    sudo apk add go-c8y-cli
    ```

4. Install go-c8y-cli helper functions and reload your shell

    ```bash
    c8y cli install
    ```

4. Go to the [Getting started section](/docs/gettingstarted/) for the next steps

---

### MacOS/Linux (Homebrew)

**go-c8y-cli** can be installed using [homebrew](https://brew.sh/) on either macOS or Linux.

1. Install **go-c8y-cli** from the custom tap

    ```bash
    brew install reubenmiller/go-c8y-cli/go-c8y-cli
    ```

    :::note
    Run the following if you want to update to the latest version (if not already installed)

    ```bash
    brew upgrade go-c8y-cli
    ```
    :::

2. Install the **c8y** helper functions and reload your shell

    ```bash
    c8y cli install
    ```

3. Go to the [Getting started section](/docs/gettingstarted/) for the next steps

:::note
You can also view the instructions on how to source the relevant plugin via the command `brew info go-c8y-cli`
:::

---

## Other recommended tools

### jq

Since the output of **go-c8y-cli** is mainly json, it is highly recommended that you install the json cli tool `jq` to help formatting the output.

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
