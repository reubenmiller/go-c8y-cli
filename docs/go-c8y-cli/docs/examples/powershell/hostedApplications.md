---
layout: default
category: Examples - Powershell
title: Applications - Web (Hosted)
---

### Hosted (web) Applications

1. Create an application placeholder for the application

    ```powershell
    New-HostedApplication -Name "myApp"
    ```

2. Once the application is finished, then it can be deployed to a staging area so it can be tested how it performs when it is hosted in the platform.

    ```powershell
    Update-HostedApplication -File "myApp.zip"
    ```

    If you don't want to active the application immediately, then you can skip the activation step by using the `-SkipActivation`

    ```powershell
    Update-HostedApplication -File "myApp.zip" -SkipActivation
    ```

3. Once the testing has finished and the application is ready to be deployed to production, then you can create, uploaded and activate the application using the following command:

    ```powershell
    New-HostedApplication -File "myApp.zip"
    ```
