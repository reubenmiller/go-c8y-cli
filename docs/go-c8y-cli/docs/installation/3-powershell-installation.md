---
title: PowerShell
---
## PowerShell (native c8y binary)

1. Clone the addons repository containing the install script, views and some example templates.

    ```sh
    git clone https://github.com/reubenmiller/go-c8y-cli-addons.git $HOME/.go-c8y-cli
    ```

    :::info
    This repository is not strictly required to run `c8y` however it makes is much more useful as it provides some  default views and templates to get the most out of `c8y`. The defaults also show you how you can create your own custom views.
    :::

2. Install go-c8y-cli binary

    ```sh
    $HOME/.go-c8y-cli/install.ps1
    ```

    :::tip
    If the following warning is displayed when you try to run the `install.ps1` script

    ```powershell
    File C:\Users\myuser\.go-c8y-cli\install.ps1 cannot be loaded because running scripts is disabled on this system...
    ```

    Then, open up a new powershell console but set the ExecutionPolicy to remotesigned (or bypass), then you can all the run the script again.

    ```powershell
    powershell -ExecutionPolicy remotesigned
    ```
    :::

3. Verify that the `c8y` binary is executable and can be found on the command line 

    ```bash
    Get-Command c8y
    ```

    ```text title="Output"
    CommandType    Name       Version    Source
    -----------    ----       -------    ------
    Application    c8y.exe    0.0.0.0    C:\Users\myusername\bin\c8y.exe
    ```

    :::tip
    Try closing your console and re-opening it so you can be sure that your setup will work next time
    :::

3. Now go to the [Getting started](/docs/gettingstarted/) section for instructions how to use it

## Alternative installation methods

:::caution
PSc8y is no longer required to get the best out of go-c8y-cli, so it recommended to use the native c8y binary instead.
:::

### Prerequisites

PowerShell (Core) 7 is available on many operating systems (i.e. Windows, MacOS, Linux). Following the [installation guide](https://docs.microsoft.com/en-us/powershell/scripting/install/installing-powershell) to install it on your machine.

:::caution
`PSc8y` no longer supports PowerShell 5. Users should install PowerShell 7 (aka pwsh) as it provides a lot of benefits and is also supported on multiple platforms (linux, MacOS and Windows).

If you still don't want to or can't install PowerShell 7 then install the native go-c8y-cli binary.
:::

### Installing PSc8y

1. Open a PowerShell (pwsh) console

    ```bash
    pwsh
    ```

    :::info Windows Users
    If you get an error regarding the `Execution policy` when installing or importing PSc8y, then you will have to start a new powershell disabling the execution policy.

    ```bash
    pwsh -ExecutionPolicy bypass
    ```
    
    More information on PowerShell's execution policy can be found on the [Microsoft website](https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_execution_policies)
    :::

1. Install `PSc8y` module from [PSGallery](https://www.powershellgallery.com/packages/PSc8y)

    ```powershell
    Install-Module PSc8y -AllowPrerelease -Repository PSGallery -AllowClobber -Scope CurrentUser
    ```

1. Import the module

    ```powershell
    Import-Module PSc8y
    ```

1. Now go to the [Getting started](/docs/gettingstarted/) section for instructions how to use it


### Updating PSc8y

Once the `PSc8y` PowerShell module has been installed, then it can be updated from within PowerShell itself assuming it was installed as per the "Installing PSc8y" section above.

1. Update to the latest version

    ```powershell
    Update-Module PSc8y
    ```

1. Import the updated module

    ```powershell
    Import-Module PSc8y -Force
    ```

1. Check the new version

    ```powershell
    Get-Module PSc8y
    ```

    ```powershell
    ModuleType Version    Name    ExportedCommands
    ---------- -------    ----    ----------------
    Script     2.0.0      PSc8y   {Add-PowershellType, Add-...
    ```

    :::info
    The version number does not show the pre-release version information.
    :::

### Downloading using Save-Module

The `PSc8y` module can be downloaded manually using `Save-Module`. This has the advantage over  `Install-Module` as you can control where the module is saved to.

:::info
`Install-Module` does not allow you to control where the modules are installed. By default the target folder is inside your home directory, however if you have your home folder synced automatically to Microsoft OneDrive then you might run into problems as `PSc8y` contains executables (i.e. `.exe` files).
:::

1. Create a folder where you want to store `PSc8y`

    ```powershell
    mkdir ~/PSModules
    ```

1. Download the `PSc8y` Module to the folder `~/PSModules`

    ```powershell
    Save-Module -Name PSc8y -AllowPrerelease -Repository PSGallery -Path ~/PSModules
    ```

    :::tip
    Save-Module also accepts `-RequiredVersion` where you can specify an exact version instead of the latest.
    :::

1. Import `PSc8y`

    ```powershell
    Import-Module ~/PSModules/PSc8y -Force
    ```

    :::tip
    You can the `~/PSModules` folder to the `$env:PSModulePath` variable so that you can use `Import-Module` without having to specify the module location each time.

    The best way to do this is by adding the following to your PowerShell profile. This can be open by editing the path stored in the `$PROFILE` variable. If the file does not exist, just create it in the specified location.

    ```powershell
    $env:PSModulePath = (Resolve-Path "~/PSModules").Path + ";" + $env:PSModulePath
    ```
    :::
