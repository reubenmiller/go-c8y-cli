---
category: Concepts - Extensions - API based commands
title: Tab completion
---

Tab completion is the killer feature on the commands line. It improves the useability of the extension and can reduce the reliance on documentation (though it shouldn't be a substitute for good docs :wink:)

The following sections detail the different tab completion mechanisms available for use.

## Tab completion with simple statics lists

If a flag has a fixed number of allowed values, then the `validationSet` option is the perfect fit. The options will be presented to the user when they try tab completion when using the flag.

For example, Cumulocity IoT's operation endpoint supports filtering the type of operations by status, and the states should be one of; `PENDING`, `EXECUTING`, `FAILED` or `SUCCESSFUL`. So using this knowledge a flag can be added to the `list` command which sets the `status` query parameter, and the users can use tab completion to check which options are available for usage. Below shows a snippet of the commands:

```yaml
commands:
  - name: list
    description: Get collection of operations
    method: GET
    path: devicecontrol/operations
    queryParameters:
      - name: status
        type: string
        description: Filter by status
        validationSet:
          - PENDING
          - EXECUTING
          - FAILED
          - SUCCESSFUL
```

Example usage

<CodeExample>

```sh
c8y organizer assets list --status <TAB><TAB>
```

</CodeExample>

```bash title="Output"
PENDING
EXECUTING
FAILED
SUCCESSFUL
```

## Tab completion using shell commands

The external tab completion is not just limited to `c8y` commands, you can also use a shell to call any commands you would like.

:::warning
If the extensions uses a shell to execute the completion command then it might make your extension less portable as Windows users might not have access to a bash shell.

Instead try to use the `c8y` command instead of calling a shell if possible.
:::

The command snippet below shows an example of a `create` command which accepts a type flag when building the request body. The type flag has a tab completion command which queries existing devices in the tenant, and returns a unique list of device types (only based on the first 2000 devices found).

```yaml
commands:
  - name: create
    description: Create managed object
    method: POST
    path: inventory/managedObjects
    body:
      - name: type
        type: string
        property: type
        description: Device type
        completion:
          type: external
          command:
            - /bin/bash
            - -c
            - c8y devices list --pageSize 2000 --select type,type --output completion | sort | uniq
```

The above command will provide the following experience when the user tries to TAB completion the value provided to the `--type` flag.

<CodeExample>

```sh
c8y organizer assets create --type <TAB><TAB>
```

</CodeExample>

```bash title="Output"
Connex Spot Monitor          -- type: Connex Spot Monitor
PMP                          -- type: PMP
RV700                        -- type: RV700
c8y_OPCUA_Device_Agent       -- type: c8y_OPCUA_Device_Agent
c8y_SNMP                     -- type: c8y_SNMP
c8y_dm_example_device        -- type: c8y_dm_example_device
c8y_lwm2m_connector_device   -- type: c8y_lwm2m_connector_device
```

:::note
The `--select type,type` part of the completion command is not a typo. By included more than one columns of data means that the the other colums will be used in the description of the option. In zsh it makes the list a bit more readable.
:::
