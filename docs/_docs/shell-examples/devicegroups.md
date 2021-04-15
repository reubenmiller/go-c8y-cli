---
layout: default
category: Examples
title: Shell-DeviceGroups
---

### Assign devices from a query to a group

```sh
c8y inventory find --query "agentVersion eq '1.*' and not(bygroupid(123456))" --includeAll | c8y devicegroups assignDevice --group 123456 --workers 2 --progress
```

### Assign devices from a file to a group

Assuming that you get handed a list just containing a list of unique names or external ids called `list.txt`:

```text
UNIQUE_ID_1
UNIQUE_ID_2
UNIQUE_ID_3
UNIQUE_ID_4
```

1. Lookup the c8y id by using a text search (this is case insensitive)

    ```sh
    cat list.txt | c8y inventory findByText --fragmentType c8y_IsDevice --select id,name,c8y_Hardware.serialNumber -o csv > list.ids.csv
    ```

    *Example output: list.ids.csv*

    ```sh
    1111,MyActualName1,UNIQUE_ID_1
    2222,MyActualName2,UNIQUE_ID_2
    3333,MyActualName3,UNIQUE_ID_3
    4444,MyActualName4,UNIQUE_ID_4
    ```

2. Check the file to verify that everything looks ok

    ```sh
    cat list.ids.csv | more
    ```

3. Assign the devices to a single static group

    ```sh
    cat list.ids.csv  | c8y devicegroups assignDevice --group 1234 --workers 2 --progress --silentStatusCodes 409
    ```

    `--silentStatusCodes` is used to silent 409 (duplicate) errors, as we don't care if the device is already assigned to the group.

    **Note**

    cat list.ids.csv  | cut -d, -f1 | c8y devicegroups assignDevice --group 1234 --workers 2 --progress --silentStatusCodes 409

    If the piped input is in a csv format, then go-c8y-cli will automatically extract the first field/column and use it to lookup the device. If the first field contains an id (i.e. just numbers), then the value will be used directly as an ID, however if it contains non ID characters (i.e. any character does not match [0-9]) then it will be treated as a name, and an additional lookup by name will be performed in order to get the device id (this will take longer as it performs one request per name).

    If store you data in csv format, it is best practice to store the id as the first field making the data more pipe friendly. However you can still select the ids using the linux command `cut`.

    ```sh
    # MyActualName1,1111,UNIQUE_ID_1
    cat list.ids.csv  | cut -d, -f2 | c8y ....
    ```
    
    Alternatively you can also store the data in a json line format (using `--output json`), then the column order does not matter as the `go-c8y-cli` is smart enough to the pluck the id value if it is present.

    **JSON line output**

    ```json
    {"name": "MyActualName1", "id": "1111", "c8y_Hardware": {"serialNumber": "UNIQUE_ID_1"}}
    {"name": "MyActualName2", "id": "2222", "c8y_Hardware": {"serialNumber": "UNIQUE_ID_2"}}
    {"name": "MyActualName3", "id": "3333", "c8y_Hardware": {"serialNumber": "UNIQUE_ID_3"}}
    {"name": "MyActualName4", "id": "4444", "c8y_Hardware": {"serialNumber": "UNIQUE_ID_4"}}    
    ```

### Assign devices where the operation was set to FAILED in a bulk operation

```sh
group=$( c8y devicegroups create --name "my_custom_group" --output csv --select id )

c8y operations list --bulkOperationId 8 --status FAILED --includeAll | c8y devicegroups assignDevice --group $group --workers 2 --progress --silentStatusCodes 409
```

### Unassign devices from a group which match a custom inventory query

The Cumulocity inventory query language supports a the `bygroupid(x)` operator which checks if a device belongs to a group. This allows a custom query to be run without needing to first get a list of devices that are already assigned to the group. 

```sh
c8y inventory find --query "agentVersion eq '1.*' and bygroupid(123456)" --includeAll | c8y devicegroups unassignDevice --group 123456 --workers 2 --progress
```

*Output*

```json
elapsed 01:34   (started: 2021-04-13T11:27:36+02:00)  ⠸ total requests sent:  399
worker 1:     ⠸                                         avg: 2.149 request/s
worker 2:     ⠸                                         avg: 2.099 request/s
```
