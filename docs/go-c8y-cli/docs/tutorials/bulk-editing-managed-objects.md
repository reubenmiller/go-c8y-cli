---
title: Bulk editing managed objects
---

import CodeExample from '@site/src/components/CodeExample';

### Example: Remove prefix for all device

**Scenario**

Devices have been registered in Cumulocity, however there was a naming error, and some devices have an incorrect prefix "MQTT Device " included in the `.name` property on the device managed object.

**Goal**

Rename incorrect devices be removing the prefix from the name.

**Procedure**

1. Find the devices which the incorrect prefix in their names

    <CodeExample>

    ```bash
    c8y devices list --name "MQTT*"
    ```

    </CodeExample>

    You can use additional fields to narrow down your search (if required)

    <CodeExample>

    ```bash
    c8y devices list --name "MQTT*" --query "type eq 'my_Custom_type'"
    ```

    </CodeExample>

2. Check that the number of matches (devices) is what you expect

    <CodeExample>

    ```bash
    c8y devices list --name "MQTT*" --withTotalPages --pageSize 1
    ```

    </CodeExample>

    The total number of devices can be read from the `totalPages` property.

    Verifying the results is a very important step, and time should be taken to check if the results does not include any other unexpected results.

3. Create your update command which will update the name of the device

    Now that you have refined the data that you want to edit, but don't execute the command just yet. This is where the `Dry` parameter is very useful!

    Since we are not just setting the name of the device, we can use a template which references each input value received via the pipeline (stdin). The template can use an in-built jsonnet function to remove the unwanted prefix from the name.  

    <CodeExample>

    ```bash
    c8y devices list --name "MQTT*" -p 1 |
        c8y devices update \
            --template "{name: std.strReplace(input.value.name, 'MQTT Device ', '')}" \
            --dry
    ```

    </CodeExample>

4. Do a test run on one device (use the page size to limit the scope)

    <CodeExample>

    ```bash
    c8y devices list --name "MQTT*" -p 1 |
        c8y devices update \
            --template "{name: std.strReplace(input.value.name, 'MQTT Device ', '')}"
    ```
    </CodeExample>

    Since we are not using the `force` parameter, we will still be prompted before each change. This protects you against accidentally executing the command for all of the result set if you forgot to change the `pageSize` parameter to 1.

5. When you are happy with your command, then you can execute it on each device in the query and use the `force` parameter

    <CodeExample>

    ```bash
    c8y devices list --name "MQTT*" -p 400 |
        c8y devices update \
            --template "{name: std.strReplace(input.value.name, 'MQTT Device ', '')}"
    ```

    </CodeExample>

6. Verify that you no longer have any devices which match your original query.

    <CodeExample>

    ```bash
    c8y devices list --name "MQTT*" --withTotalPages -p 1
    ```

    </CodeExample>
