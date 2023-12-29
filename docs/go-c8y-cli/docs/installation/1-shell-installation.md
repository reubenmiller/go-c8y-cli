---
title: Shell
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Installation

`go-c8y-cli` is available as a pre-built linux package which can be installed and updated via a package manager. Please following the instructions in your operating system.

It is recommended to install `go-c8y-cli` using a package manager as it makes it easier to update it in the future, and it will be available for all users.

After the installation, follow the instructions to [setup your shell profile](/docs/installation/shell-installation#setting-up-your-shell-profile).

### Debian / Ubuntu (apt)

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

4. Follow the instructions to [setup your shell profile](/docs/installation/shell-installation#setting-up-your-shell-profile)

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

2. Update the repo then install/update `go-c8y-cli`

    ```sh
    sudo dnf update
    sudo dnf install go-c8y-cli
    ```

3. Follow the instructions to [setup your shell profile](/docs/installation/shell-installation#setting-up-your-shell-profile)

:::note
You can install `go-c8y-cli` via `yum` by just replacing `dnf` with `yum` in the above commands.
:::

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

3. Update the repo then install/update `go-c8y-cli`

    ```bash
    sudo apk update
    sudo apk add go-c8y-cli
    ```

4. Follow the instructions to [setup your shell profile](/docs/installation/shell-installation#setting-up-your-shell-profile)


### MacOS/Linux (Homebrew)

`go-c8y-cli` can be installed using [homebrew](https://brew.sh/) on either macOS or linux.

1. Add the tap

    ```bash
    brew tap reubenmiller/go-c8y-cli
    ```

2. Update brew then install `go-c8y-cli`

    ```bash
    brew update
    brew install go-c8y-cli
    ```

    :::note
    Run the following if you want to update to the latest version (if not already installed)

    ```bash
    brew upgrade go-c8y-cli
    ```
    :::

3. Edit your preferred shell by executing snippet (it will import functions each time you load your shell)

    <Tabs
    groupId="shell-types"
    defaultValue="bash"
    values={[
        { label: 'Bash', value: 'bash', },
        { label: 'Zsh', value: 'zsh', },
        { label: 'Fish', value: 'fish', },
        { label: 'PowerShell', value: 'powershell', },
    ]
    }>
    <TabItem value="bash">

    ```bash
    echo 'source "$(brew --prefix)/etc/go-c8y-cli/shell/c8y.plugin.sh"' >> ~/.bashrc
    ```

    </TabItem>
    <TabItem value="zsh">

    ```bash
    echo 'source "$(brew --prefix)/etc/go-c8y-cli/shell/c8y.plugin.zsh"' >> ~/.zshrc
    ```

    </TabItem>
    <TabItem value="fish">

    ```bash
    mkdir -p ~/.config/fish
    echo 'source "$(brew --prefix)/etc/go-c8y-cli/shell/c8y.plugin.fish"' >> ~/.config/fish/config.fish
    ```

    </TabItem>

    <TabItem value="powershell">

    ```powershell
    New-Item -type directory -path ~/.config/powershell -Force
    '. "$(brew --prefix)/etc/go-c8y-cli/shell/c8y.plugin.ps1"' >> ~/.config/powershell/Microsoft.PowerShell_profile.ps1
    ```

    </TabItem>

    </Tabs>

4. Restart your shell to reload your profile

:::note
You can also view the instructions on how to source the relevant plugin via the command `brew info go-c8y-cli`
:::

### Manually (via script)

:::info
The install script currently requires you to have `jq` installed. `jq` is a cli json parsing tool which is recommended to have anyways to do complex json manipulations that you might need during your daily use of `c8y`.

See the [jq website](https://stedolan.github.io/jq/download/) for details how to install it on your operating system.
:::

`go-c8y-cli` can also be installed by cloning a git repository and running an install script. It will install the latest binary and add the plugin script to your shell profile.

This method does not require sudo rights, however the binary will be located inside your user's home folder.

1. Clone the addons repository containing the install script, views and some example templates.

    ```bash
    git clone https://github.com/reubenmiller/go-c8y-cli-addons.git ~/.go-c8y-cli
    ```

2. Install go-c8y-cli binary

    ```bash
    ~/.go-c8y-cli/install.sh
    ```

3. Add the installation path to your `PATH` variable (ideally in your shell profile)

    <Tabs
    groupId="shell-types"
    defaultValue="bash"
    values={[
        { label: 'Bash', value: 'bash', },
        { label: 'Zsh', value: 'zsh', },
        { label: 'Fish', value: 'fish', },
    ]
    }>
    <TabItem value="bash">

    ```bash title="file: ~/.bashrc"
    export PATH="$HOME/bin:$PATH"
    ```

    </TabItem>
    <TabItem value="zsh">

    ```bash title="file: ~/.zshrc"
    export PATH="$HOME/bin:$PATH"
    ```

    </TabItem>
    <TabItem value="fish">

    ```bash title="file: ~/.config/fish/config.fish"
    set -gx PATH "$HOME/bin:$PATH"
    ```

    </TabItem>
    </Tabs>

    Reload your shell, or source your profile directly, i.e. bash: `source ~/.bashrc`

4. Verify that the `c8y` binary is executable and can be found on the command line 

    ```bash
    which c8y
    ```

    ```text title="Output"
    /home/myuser/bin/c8y
    ```

    :::tip
    Try closing your console and re-opening it so you can be sure that your setup will work next time
    :::

---

## Setting up your shell profile

Add the following line to your shell profile to enable the shell functions like `set-session` and to configure tab completion for `go-c8y-cli`.

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Bash', value: 'bash', },
    { label: 'Zsh', value: 'zsh', },
    { label: 'Fish', value: 'fish', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash title="file: ~/.bashrc"
source "/etc/go-c8y-cli/shell/c8y.plugin.sh"

# or if you installed it via the script
source "$HOME/.go-c8y-cli/shell/c8y.plugin.sh"
```

</TabItem>
<TabItem value="zsh">

```bash title="file: ~/.zshrc"
source "/etc/go-c8y-cli/shell/c8y.plugin.zsh"

# or if you installed it via the script
source "$HOME/.go-c8y-cli/shell/c8y.plugin.zsh"
```

</TabItem>
<TabItem value="fish">

```bash title="file: ~/.config/fish/config.fish"
source "/etc/go-c8y-cli/shell/c8y.plugin.fish"

# or if you installed it via the script
source "$HOME/.go-c8y-cli/shell/c8y.plugin.fish"
```

</TabItem>
<TabItem value="powershell">

```powershell title="file: ~/.config/powershell/Microsoft.PowerShell_profile.ps1"
. "/etc/go-c8y-cli/shell/c8y.plugin.ps1"

# or if you installed it via the script
. "$HOME/.go-c8y-cli/shell/c8y.plugin.ps1"
```

</TabItem>
</Tabs>

:::note
If you don't import the plugin script, then you will have to use the following to set your session:

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Bash', value: 'bash', },
    { label: 'Zsh', value: 'zsh', },
    { label: 'Fish', value: 'fish', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
eval $(c8y sessions set --shell bash)
```

</TabItem>
<TabItem value="zsh">

```bash
eval $(c8y sessions set --shell zsh)
```

</TabItem>
<TabItem value="fish">

```bash
c8y sessions set --shell fish | source
```

</TabItem>
<TabItem value="powershell">

```powershell
c8y sessions set --shell powershell | out-string | Invoke-Expression
```

</TabItem>
</Tabs>

:::


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

## Getting started

After `go-c8y-cli` has been installed, follow the [Getting started](/docs/gettingstarted/) section for instructions how to use it.
