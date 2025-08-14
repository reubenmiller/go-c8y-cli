---
category: Installation
title: Installation
id: installation
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import DocCardList from '@theme/DocCardList';

%%c8y%% can be installed on pretty much any host you can think of (Linux, macOS, Windows etc.), and it consists of a single binary called `c8y` (or `c8y.exe` for Windows users).

For convenience, there is a script that can be used to install %%c8y%% which will automatically download and install the current version for you. If you don't want to use the convenience script, then follow the Alternative instructions.


### Linux / macOS / WSL2

Install %%c8y%% on Linux, macOS or WSL2 (Windows Subsystem for Linux) using the following one-liner.

<Tabs
    groupId="host_os"
    defaultValue="wget"
    values={[
        { label: 'wget', value: 'wget' },
        { label: 'curl', value: 'curl' },
    ]
    }>
    <TabItem value="wget">

    ```bash
    wget -qO - https://goc8ycli.netlify.app/install.sh | sh -s
    ```

    </TabItem>
    <TabItem value="curl">

    ```bash
    curl -fsSL https://goc8ycli.netlify.app/install.sh | sh -s
    ```

    </TabItem>
</Tabs>

The convenience script can be customized, for example you can choose how it is installed using the `-p <package_manager>` flag. To see the full installation options, then just pass the `--help` to the script.

<Tabs
    groupId="host_os"
    defaultValue="wget"
    values={[
        { label: 'wget', value: 'wget' },
        { label: 'curl', value: 'curl' },
    ]
    }>
    <TabItem value="wget">

    ```bash
    wget -qO - https://goc8ycli.netlify.app/install.sh | sh -s -- --help
    ```

    </TabItem>
    <TabItem value="curl">

    ```bash
    curl -fsSL https://goc8ycli.netlify.app/install.sh | sh -s -- --help
    ```

    </TabItem>
</Tabs>

### PowerShell

%%c8y%% can be installed using PowerShell using the following one-liner.

```bash
. { Invoke-WebRequest https://goc8ycli.netlify.app/install.ps1 } | Invoke-Expression; install-c8y
```

### Alternative Installation Methods

If you don't want to use the convenience script, then follow any of the alternative installation methods.

<DocCardList />

:::note
If you are having problems installing %%c8y%%, or have some suggestions to improve it, then please [create an issue](https://github.com/reubenmiller/go-c8y-cli/issues/new) to address the problem.

If you can't install it using any of the described methods then you can just download the binary from the [Releases Page](https://github.com/reubenmiller/go-c8y-cli/releases).
:::
