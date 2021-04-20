---
layout: default
category: Tutorials - Powershell
title: Find event occurences across devices
---

### Example: Find which devices have sent a specific event in the last hour

**Scenario**

A user notices an event in the Cumulocity UI which related to an error condition. They would like to know how many other devices have also sent this event recently, to know how wide-spread the problem is.

**Goal**

Get a list of devices where a specific event has been sent to Cumulocity in the last hour.


**Procedure**

1. Get the event type of the event that you want to search for.

    In this example we searching for the type `device_disconnected`.


2. Find how many events with this type `device_disconnected` were created in the last 1 hour

    ```
    $events = Get-EventCollection -Type device_disconnected -DateFrom "-1h" -PageSize 1 -WithTotalPages
    ```

    ```powershell
        self: https://{tenant}.{url}/event/events?dateFrom=2020-09-30T13:31:08.357206%2B02:00&type=device_disconnected&withTotalPages=true&pageSize=1&currentPage=1
        next: https://{tenant}.{url}/event/events?dateFrom=2020-09-30T13:31:08.357206%2B02:00&type=device_disconnected&withTotalPages=true&pageSize=1&currentPage=2

    currentPage     pageSize        totalPages      events
    -----------     --------        ----------      ------
    1               1               245             ....
    ```

    The `totalPages` property on the console show the total number of events found using the search criteria. However some of these events may have come from the same devices.

    To get a list of the unique devices, in-built powershell functions can be used.

    We will set the `PageSize` parameter to something higher than the `totalPages` response, so that we can be sure we have all the result, then we assign the result back to a variable.

    We use the assigned results in `$events` and use dot notation to reference the `.source.id` property of every item in the `$events` array. This result is then piped to the `Sort-Object` cmdlet which removes

    ```powershell
    $events = Get-EventCollection -Type device_disconnected -DateFrom "-1h" -PageSize 2000
    
    # Get unique list of ids
    $events.source.id | Sort-Object -Unique | Measure-Object
    ```

    *Console output*

    ```powershell
    Count             : 147
    Average           :
    Sum               :
    Maximum           :
    Minimum           :
    StandardDeviation :
    Property          :
    ```

    The results show that 147 devices have sent this event in the last hour.

    If you want the ids or names of the devices then just leave out the last pipe `| Measure-Object`.

    ```
    # Get unique list of names
    $events.source.name | Sort-Object -Unique

    # Get unique list of ids
    $events.source.id | Sort-Object -Unique
    ```

    **Notes**

    * The events API is restricted to a max page size of 2000 events
