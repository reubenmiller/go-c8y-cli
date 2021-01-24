---
layout: default
category: Getting started
title: Powershell
---

1. `PSc8y` should already be installed. If not, then following the [install guide](https://reubenmiller.github.io/go-c8y-cli/docs/1-powershell-installation/)

1. Import the PSc8y module into your Powershell console

    ```powershell
    Import-Module PSc8y
    ```

    **Note for Windows**

    If you get an error regarding the `Execution policy` when installing or importing PSc8y, then you will have to start a new powershell conc

    ```sh
    pwsh -ExecutionPolicy bypass
    ```
    
    More information on PowerShell's execution policy can be found on the [Microsoft website](https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_execution_policies)

1. Create a new session (containing your Cumulocity credentials, host and tenant)

    ```sh
    New-Session -Name "custom-session-name" -Host "https://example.eu-latest.cumulocity.com" -Tenant "t123456789"
    ```

    You will be prompted for your username and password. Alternatively you can also enter the credentials using the `-Credential` parameter.

    The `-Name` option allows use to specify your own name for the session which will help you identify it when using it in the future. The command create a json file using the following format: `~/.cumulocity/<Name>.json`

    **Alternative: Create session manually using a text editor**

    1. Create the `~/.cumulocity` folder where the session files will be stored. If this folder already exists, then skip this step.

        ```sh
        New-Item -Type Directory -Path "~/.cumulocity" -ErrorAction ignore
        ```

    1. Create a new session file in the `~/.cumulocity` folder
    
        ```sh
        New-item -Type File -Path "~/.cumulocity/custom-session-name.json"
        ```

    1. Update the newly created file (in this example `~/.cumulocity/custom-session-name.json`) with the following content and editing it the host, tenant, username and password.

        ```json
        {
            "host": "https://example.eu-latest.cumulocity.com",
            "tenant": "t123456789",
            "username": "myUserName@example.com",
            "password": "hopefully-something-complicated",
            "description": "",
            "useTenantPrefix": true,
            "microserviceAliases": {}
        }
        ```

        **Warning**

        Please use the `tenantId` (i.e. t12345678) in the `tenant` property if you have a new t-style tenant id.

    **Tips:**

    Once you have an existing session file in `~/.cumulocity/`, it is easy to create a new session by copying an existing file, rename it, and updating the Cumulocity properties.
    
    The cli tool will scan the `~/.cumulocity/` folder for `.json` files when calling `Set-Session`.
    
    The cli tool reads the .json session file every time a command is run, so if you any session details (like password etc.), then the cli tool will automatically use the updated settings (you don't need to run `Set-Session` again).


1. Activate the session using the interactive session selector

    ```sh
    Set-Session
    ```

1. Test your credentials by getting your current user information from the platform

    ```sh
    Get-CurrentUser
    ```

    **Note**

    If your credentials are incorrect, then you can update the session file stored in the `~/.cumulocity` directory

1. Now you're ready to go. You can get a list of available commands by using help menu

    ```sh
    # List commands
    Get-Command -Module PSc8y

    # Get help for a command
    Get-Help Get-DeviceCollection -Full
    ```

### Switching sessions

The sessions can be changed again by using the interactive session selector

```sh
Set-Session
```
