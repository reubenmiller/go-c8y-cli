---
layout: powershell
category: Tutorials - Powershell
title: Extend using module
---

:::tip
It is recommended to use [extensions](/docs/concepts/extensions/) instead of creating your own PowerShell module. Extensions are available in go-c8y-cli >= 2.30.0 and provide a more portable/native experience. They are just as easy to install and share, and can be used in any shell.

Or just check out the [Tutorials](/docs/tutorials/extensions/).
:::

An example how to extend the `PSc8y` PowerShell module using another PowerShell module is shown in the following demo project:

[Example PSc8y.example Module](https://github.com/reubenmiller/PSc8y.example)

The full instructions can be viewed from the above project link, however the workflow for using the project would be as follows:


1. Open a PowerShell (pwsh) console
    ```bash
    pwsh

    # Or if you are using PowerShell 5.1 on windows
    powershell
    ```

2. Clone the project (requires git)

    ```bash
    git clone https://github.com/reubenmiller/PSc8y.example.git
    cd PSc8y.example
    ```

3. Import the PowerShell module

    ```powershell
    Import-Module ./ -Force
    ```

    **Note**: The `-Force` parameter is important when re-importing a module from a directory. If it is not used and the module has already been imported once, then any new functions or changes will not be loaded!

4. Show a list of the commands in the module

    ```powershell
    Get-Command -Module PSc8y.example
    ```

5. Get help for a specific module

    ```powershell
    Get-Help Clear-OperationCollection -Full
    ```

---

## Adding your own functions

1. Add your own PowerShell functions into the `public` folder. 

    **Note:** The name of the function needs to be the same as the file name!

    For example: If you add a file called `Get-MyDeviceCollection.ps1` into the `public` folder, then the file should contain the following contents:

    ```powershell
    Function Get-MyDeviceCollection {
        # insert your function body here

    }
    ```

2. Re-Import the module into your PowerShell session

    ```powershell
    Import-Module ./ -Force
    ```

3. Use your new function

---

## Creating a new PowerShell module from the example project

The module can be used as a template for your own PSc8y extension module. For convenience there is a script to rename the module and be used as follows:

1. Run the rename module script

    ```powershell
    ./scripts/Invoke-RenameModule.ps1 -Name "PSc8y.mymodule" -OutputFolder "../"
    ```

    **Notes**
    
    You can change the -Name parameter to anything you want. By convention the module name should start with `PSc8y.` to show that the module is an extension of `PSc8y`, however this convention is not enforced.

2. Change directory to the new module output folder, and import it into your current PowerShell session

    ```powershell
    cd "../PSc8y.mymodule"
    Import-Module ./ -Force
    ```

---

## Importing your custom module automatically when loading PowerShell

If you want to import your module automatically, then you can add the `Import-Module` statement into your profile.ps1 file. The instructions below show how to do this:

1. Find out where your PowerShell profile file is 

    ```powershell
    $PROFILE
    ```

2. Edit the file in an editor.

    ```powershell
    # If you using VS Code, open it with
    code $PROFILE
    ```

    Or just open the path from step 1 in your preferred text editor.

3. Add the following import statement

    ```powershell
    # import custom module (please change the path to the relevant path of your module!!!)
    Import-Module "$Env:HOME/path/to/module/PSc8y.mymodule" -Force
    ```

    You need to pass the path to the folder where you saved your module.

4. Start a new PowerShell session

5. Use a command from your module
