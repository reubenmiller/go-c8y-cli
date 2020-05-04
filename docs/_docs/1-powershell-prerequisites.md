---
layout: default
category: Installation
order: 199
title: PowerShell Prerequisites
---


Due to ongoing changes to the PowerShell loading module, we first need to make sure some inbuilt PowerShell modules are up-to-date, otherwise you will have problems trying to install the [PSc8y](https://www.powershellgallery.com/packages/PSc8y) module from the PowerShell Gallery (online dotnet/powershell community repository).

The instructions slightly differ depending on which operating system you are using (i.e. Windows, MacOS or Linux). However the good news is that you only have to install the prerequisites once.


##### Windows

1. Start an PowerShell Console as Administrator

1. Check that you have a recent version of PowerShell (version 5.0 or higher)

    ```sh
    $PSVersionTable.PSVersion.Major
    ```

    If you have an older version of PowerShell then please update it, or install [PowerShell 6 (Core)](https://github.com/PowerShell/PowerShell/releases) as it can run along side PowerShell 5. A portable zip version is also available.

1. Install Latest version from PowerShell Gallery and the PowerShellGet Module

    To get the latest version from PowerShell Gallery, you should first install the latest Nuget provider. You will need to run PowerShell as an Administrator for all  the following commands

    ```sh
    Set-ExecutionPolicy RemoteSigned -Force;

    Install-PackageProvider Nuget –Force;
    Install-Module –Name PowerShellGet –Force;

    Exit
    ```

    If the `Install-Module` command fails, then try the command gain adding the `-AllowClobber` option.


    Note:

    In the future the library can be updated by using a single command

    ```sh
    Update-Module -Name PowerShellGet
    Exit
    ```

    If there are any errors then please read the original [blog](https://www.thomasmaurer.ch/2019/02/update-powershellget-and-packagemanagement/) where these instructions were copied from

##### MacOS and Linux

    1. Install the newest version of PowerShellGet

    ```sh
    Install-Module –Name PowerShellGet –Force;
    ```

    1. Close the powershell session. Unfortunately you need to reload the console for this step.

    ```sh
    exit
    ```

    1. Open up a powershell console again

    ```sh
    pwsh
    ```

    Then import PowerShellGet

    ```sh
    Import-Module PowerShellGet
    ```
