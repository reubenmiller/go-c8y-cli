---
layout: default
category: Installation - PowerShell
order: 200
title: Installation
---

Try it out in a Cloud Shell!

[![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.svg "Launch Cloud Shell"){:height="26px" width="180px"}](https://ssh.cloud.google.com/cloudshell/editor?shellonly=true)

[![Launch Cloud Shell](https://shell.azure.com/images/launchcloudshell.png "Launch Cloud Shell")](https://shell.azure.com/powershell)

## Installing PSc8y

### Windows

#### PowerShell 7 and newer

Powershell 7 is not installed on Windows by default, however it can be installed using the following [instructions](https://docs.microsoft.com/en-us/powershell/scripting/install/installing-powershell-core-on-windows?view=powershell-7).

Once powershell is installed, then `PSc8y` can be installed using:

1. Open a PowerShell (pwsh) console (i.e. pwsh.exe)

    ```sh
    pwsh.exe
    ```

    **Note for Windows**

    If you get an error regarding the `Execution policy` when installing or importing PSc8y, then you will have to start a new powershell conc

    ```sh
    pwsh -ExecutionPolicy bypass
    ```
    
    More information on PowerShell's execution policy can be found on the [Microsoft website](https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_execution_policies)

1. Install `PSc8y` module from [PSGallery](https://www.powershellgallery.com/packages/PSc8y)

    ```powershell
    Install-Module PSc8y -Repository PSGallery -AllowClobber -Scope CurrentUser
    ```

1. Import the module

    ```powershell
    Import-Module PSc8y
    ```

1. Now go to the [Getting started](https://reubenmiller.github.io/go-c8y-cli/docs/2-getting-started-powershell/) section for instructions how to use it

#### PowerShell 5

Powershell 5 is no longer officially supported. Users should install PowerShell 7 (aka pwsh) as it provides a lot of benefits and is also supported on multiple platforms (linux, MacOS and Windows).

Following the instructions to install [PowerShell 7 (Core)](https://github.com/PowerShell/PowerShell/releases).

### MacOS

If you do not already have PowerShell (pwsh) on your system then it can be installed on my using these [instructions](https://docs.microsoft.com/en-us/powershell/scripting/install/installing-powershell-core-on-macos?view=powershell-7).

Once powershell is installed, then `PSc8y` can be installed using:

1. Open a PowerShell Console from your Terminal app
    
    ```powershell
    pwsh
    ```

1. Install `PSc8y` module from [PSGallery](https://www.powershellgallery.com/packages/PSc8y)

    ```powershell
    Install-Module PSc8y -Repository PSGallery -AllowClobber -Scope CurrentUser
    ```

1. Import the module

    ```powershell
    Import-Module PSc8y
    ```

1. Now go to the [Getting started](https://reubenmiller.github.io/go-c8y-cli/docs/2-getting-started-powershell/) section for instructions how to use it

### Linux

If you do not already have PowerShell (pwsh) on your system then it can be installed on my using these [instructions](https://docs.microsoft.com/en-us/powershell/scripting/install/installing-powershell-core-on-linux?view=powershell-7).

Once powershell is installed, then `PSc8y` can be installed using:

1. Open a PowerShell Console
    
    ```powershell
    pwsh
    ```

1. Install `PSc8y` module from [PSGallery](https://www.powershellgallery.com/packages/PSc8y)

    ```powershell
    Install-Module PSc8y -Repository PSGallery -AllowClobber -Scope CurrentUser
    ```

1. Import the module

    ```powershell
    Import-Module PSc8y
    ```

1. Now go to the [Getting started](https://reubenmiller.github.io/go-c8y-cli/docs/2-getting-started-powershell/) section for instructions how to use it


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
    Script     1.1.0      PSc8y   {Add-PowershellType, Add-...
    ```

    **Note:**

    The version number does not show the pre-release version information.
