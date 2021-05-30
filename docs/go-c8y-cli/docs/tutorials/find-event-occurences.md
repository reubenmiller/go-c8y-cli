---
title: Find event occurrences across devices
---

import CodeExample from '@site/src/components/CodeExample';
### Example: Find which devices have sent a specific event in the last hour

**Scenario**

A user notices an event in the Cumulocity UI which related to an error condition. They would like to know how many other devices have also sent this event recently, to know how wide-spread the problem is.

**Goal**

Get a list of devices where a specific event has been sent to Cumulocity in the last hour.


**Procedure**

1. Get the event type of the event that you want to search for.

    In this example we searching for the type `device_disconnected`.


2. Find how many events with this type `device_disconnected` were created in the last 1 hour

    <CodeExample>
    
    ```bash
    c8y events list --type device_disconnected --dateFrom "-1h" -p 1 --withTotalPages
    ```

    </CodeExample>

    ```bash title="Output"
    | totalPages | pageSize   | currentPage |
    |------------|------------|-------------|
    | 245        | 1          | 1           |
    ```

    The `totalPages` property on the console show the total number of events found using the search criteria. However some of these events may have come from the same devices.

    To get a list of the unique devices, native commands can be used.

    We will set the `pageSize` parameter to something higher than the `totalPages` response, so that we can be sure we have all the result, then we assign the result back to a variable.

    We use the assigned results in `$events` and use dot notation to reference the `.source.id` property of every item in the `$events` array. This result is then piped to the `Sort-Object` cmdlet which removes

    <CodeExample>

    ```bash
    c8y events list \
        --type device_disconnected \
        --dateFrom "-1h" \
        -p 2000 \
        --select source.id -o csv | sort --unique | wc -l
    ```

    ```powershell
    $events = Get-EventCollection -Type device_disconnected -DateFrom "-1h" -PageSize 2000
    
    # Get unique list of ids
    $events.source.id | Sort-Object -Unique | Measure-Object
    ```

    </CodeExample>

    ```bash title="Output"
    147
    ```

    The results show that 147 devices have sent this event in the last hour.

    If you want the ids or names of the devices then just leave out the last pipe `| wc -l`.

    <CodeExample>

    ```bash
    c8y events list \
        --type device_disconnected \
        --dateFrom "-1h" \
        -p 2000 \
        --select "source.id,source.name" -o csv | sort --unique
    ```

    </CodeExample>

    **Notes**

    * The events API is restricted to a max page size of 2000 events, but you can use the `includeAll` parameter to fetch all of the results as go-c8y-cli will look after the paging for you.
