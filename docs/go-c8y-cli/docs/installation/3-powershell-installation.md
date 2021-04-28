---
title: PowerShell
---
## Prerequisites

Due to ongoing changes to the PowerShell loading module, we first need to make sure some inbuilt PowerShell modules are up-to-date, otherwise you will have problems trying to install the [PSc8y](https://www.powershellgallery.com/packages/PSc8y) module from the PowerShell Gallery (online dotnet/powershell community repository).

The instructions slightly differ depending on which operating system you are using (i.e. Windows, MacOS or Linux). However the good news is that you only have to install the prerequisites once.

### PowerShell 7 (Linux/MacOS/Windows)

Following the instructions to install [PowerShell 7 (Core)](https://github.com/PowerShell/PowerShell/releases).

:::caution
Powershell 5 is no longer supported. Users should install PowerShell 7 (aka pwsh) as it provides a lot of benefits and is also supported on multiple platforms (linux, MacOS and Windows).
:::

## Installing PSc8y (Linux/MacOS/Windows)

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
    Install-Module PSc8y -Repository PSGallery -AllowClobber -Scope CurrentUser
    ```

1. Import the module

    ```powershell
    Import-Module PSc8y
    ```

1. Now go to the [Getting started](../gettingstarted) section for instructions how to use it


## Updating PSc8y

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
## Manually installing PSc8y

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
    Save-Module -Name PSc8y -Repository PSGallery -Path ~/PSModules
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
