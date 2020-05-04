---
layout: default
category: Troubleshooting
title: Installation
---

### PowerShell Module installation Problems on (MacOS and Linux)

**Note:**

If you're having problems with installing PowerShell modules using `Install-Module`, then try updating to the latest PowerShellGet module directly from the source using the following:

1. Clone the PowerShellGet repository

    ```sh
    git clone https://github.com/PowerShell/PowerShellGet
    ```

1. Import the module from the cloned directory

    ```sh
    Remove-Module PowerShellGet -ErrorAction SilentlyContinue
    Import-Module PowerShellGet/src/PowerShellGet -Force
    ```
