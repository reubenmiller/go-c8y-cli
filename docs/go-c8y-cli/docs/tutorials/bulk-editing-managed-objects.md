---
layout: default
category: Tutorials - Powershell
title: Bulk editing managed objects
---

### Example: Remove prefix for all device

**Scenario**

Devices have been registered in Cumulocity, however there was a naming error, and some devices have an incorrect prefix "MQTT Device " included in the `.name` property on the device managed object.

**Goal**

Rename incorrect devices be removing the prefix from the name.

**Procedure**

1. Find the devices which the incorrect prefix in their names

    ```powershell
    devices MQTT*
    ```

    You can use additional fields to narrow down your search (if required)

    ```powershell
    devices MQTT* -Query "type eq 'my_Custom_type'"
    ```

2. Check that the number of matches (devices) is what you expect

    ```bash
    devices MQTT* -PageSize 1 -WithTotalPages
    ```

    The total number of devices can be read from the `TotalPages` property.

    Verifying the results is a very important step, and time should be taken to check if the results does not include any other unexpected results.

3. Create your update command which will update the name of the device

    Now that you have refined the data that you want to edit, but don't execute the command just yet. This is where the `-WhatIf` parameter is very useful!

    Since we are not just setting the name of the device, we actually want to update the name of the device based on the existing name, therefore we need to pipe to the `ForEach-Object` so that we can use the PowerShell `-replace` operator to remove the "MQTT Device " text from the current name.

    ```powershell
    devices MQTT* -PageSize 1 | ForEach-Object {
        Update-ManagedObject `
            -Id $_.id `
            -Data @{
                name = $_.name -replace "MQTT Device ", ""
            } -WhatIf
    }
    ```

4. Do a test run on one device (use the page size to limit the scope)

    ```powershell
    devices MQTT* -PageSize 1 | ForEach-Object {
        Update-ManagedObject -Id $_.id -Data @{
            name = $_.name -replace "MQTT Device ", ""
        }
    }
    ```

    Since we are not using the `-Force` parameter, you will still be prompted before each change. This protects you against accidentally executing the command for all of the result set if you forgot to change the `-PageSize` parameter to 1.

    If you want to view the response of the update command but forgot to pipe it to `json`, then you can use the following to show the last response.

    ```bash
    $_data | json
    ```

5. When you are happy with your command, then you can execute it on each device in the query and use the `-Force` parameter

    ```powershell
    devices MQTT* -PageSize 400 | ForEach-Object {
        Update-ManagedObject -Id $_.id -Data @{
            name = $_.name -replace "MQTT Device ", ""
        }
    }
    ```

6. Verify that you no longer have any devices which match your original query.

    ```bash
    devices MQTT* -PageSize 1 -WithTotalPages
    ```
