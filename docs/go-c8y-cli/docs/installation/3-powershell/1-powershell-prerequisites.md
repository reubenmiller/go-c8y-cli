---
layout: default
category: Installation - PowerShell
order: 199
title: Prerequisites
slug: prerequisites
---


Due to ongoing changes to the PowerShell loading module, we first need to make sure some inbuilt PowerShell modules are up-to-date, otherwise you will have problems trying to install the [PSc8y](https://www.powershellgallery.com/packages/PSc8y) module from the PowerShell Gallery (online dotnet/powershell community repository).

The instructions slightly differ depending on which operating system you are using (i.e. Windows, MacOS or Linux). However the good news is that you only have to install the prerequisites once.


##### Windows

**PowerShell 7**

Following the instructions to install [PowerShell 7 (Core)](https://github.com/PowerShell/PowerShell/releases).

**PowerShell 5**

Powershell 5 is no longer officially supported. Users should install PowerShell 7 (aka pwsh) as it provides a lot of benefits and is also supported on multiple platforms (linux, MacOS and Windows).

##### MacOS and Linux

1. Install the newest version of PowerShellGet

    ```bash
    Install-Module –Name PowerShellGet –Force;
    ```

1. Close the powershell session. Unfortunately you need to reload the console for this step.

    ```bash
    exit
    ```

1. Open up a powershell console again

    ```bash
    pwsh
    ```

1. Import PowerShellGet

    ```bash
    Import-Module PowerShellGet
    ```
