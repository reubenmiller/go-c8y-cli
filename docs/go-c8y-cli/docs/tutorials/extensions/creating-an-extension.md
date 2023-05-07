---
category: Tutorials - Extensions
title: Creating your own extension
---

## Preface

Extensions support an api which is not so dissimilar to the [OpenAPI Specification](https://swagger.io/specification/). It was by design that the command spec should not use OpenAPI spec as there is usually a useability layer on top of the API to make the command more user-friendly.

In the future there might be tooling to create the command specs automatically from the open api specs, which would make it quicker to create or update go-c8y-cli extensions.


## Example 1: Creating an extension for a microservice

Let's say that you have developed a microservice that is deployed in Cumulocity IoT, and now you would like to create some a CLI interface which can be used by users to get the most out of your new microservice.

For this example, let's assume we have a microservice called `organizer` and it is responsible for the management of some generic IoT assets.

|API|Description|Command (e.g. the one we are going to create)|
|---|-----------|---------------|
|`GET /service/organizer/assets`| Get a list of the already existing plans | `c8y organizer assets list` |
|`PUT /service/organizer/assets/{id}`| Update an existing plan | `c8y organizer assets update --id "1234"` |
|`GET /service/organizer/assets/{id}`| Get a single plan by id | `c8y organizer assets get --id "1234"` |
|`POST /service/organizer/assets`| Create a new plan | `c8y organizer assets create --template ./sometemplate.jsonnet` |
|`DELETE /service/organizer/applications/{id}`| Delete an existing plan | `c8y organizer assets delete --id "1234"` |


:::note
Though with the power of `go-c8y-cli`, the idea is that we would like to provide a similar interface for the extension so it feels and behaves like any other go-c8y-cli command. So ideally you should try to align as much as possible, for example use `create` instead of `new` when creating a new object, and use `update` instead `set` or any other synonym.
:::

In addition to the command we are going to provide some advanced functionality to make the commands more usable. We will also look at adding the following features:

* Support pipeline (using the asset id as the pipeline iterator)
* Tab completion for the asset
* Supported named lookups of assets

### Step 1: Creating the extension scaffolding

A new extension can be either created manually (if you know the required structure), or you can use the following command which creates an extension:

```sh
cd ~
c8y extensions create organizer
```

:::note
You may have noticed that the extension will have the `c8y-` prefix. This is so that it is easier to find the extensions as extensions are intended to be shared amongst users.
:::

Once created you can already install the extension so that we can experiment with our commands as we go.

```
cd c8y-organizer
c8y extensions install .
```

You should see the following message if it was installed correctly.

```
✓ Installed extension c8y-organizer
```

Then you can also checkout some of the inbuilt commands

```sh
# Show the help
c8y organizer --help

# Try out on of the command from the help (using dry)
c8y organizer devices list --dry
```

### Step 2: Inspect the overall extension structure

Let's start off by opening up the extension folder in an editor like VS Code. Assuming you are still on the console and inside the root folder of the extension folder, `c8y-organizer`, you can open up the c8y extension created in the previous step using:

```
code .
```

:::tip
Editing the yaml files is much easier in VSCode if you have the YAML extension (developer id `redhat.vscode-yaml`), as this adds tab completion support in the editor, so you don't have to keep looking up the schema yourself.
:::


Before we dive in, let's have a look at how the api commands are structured, so that we get to see what mechanisms are at work to make it easier to modify them to our use-case.

**api/devices.yaml**

The spec files under the `api/` folder are yaml based definitions which tell `go-c8y-cli` how it should build the command and what API call it should send.

The first section of the `api/devices.yaml` file controls the command group to which all commands defined in the file will be added under. The name of the file does not really matter, however it best practices to keep the file name aligned with the `group.name`.

Below shows a snippet that the commands will be included under `c8y organizer devices <subcommands>`.

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/feat/extensions-manager/tools/schema/extensionCommands.json
group:
  name: devices
  description: Manage devices
  descriptionLong: |
    More descriptive block which can even list include example of how to use different commands together.

    c8y organizer devices list | c8y c8y organizer update --name "My name"
```

The next section `command` is an array containing all of the commands.

### Step 3: Adding the get command


```yaml
commands:
  - name: get
    description: Get asset
    method: GET
    path: service/organizer/{id}    # <= It uses a variable "{id}" which is defined in `pathParameters`
    pathParameters:
      - name: id    # <= This should have the same name as placeholder in the .path field!
        type: string
        description: Asset
```

|Name|Description|
|----|-----------|
|`name`| Name of the command |
|`method`| Which HTTP method to use |
|`path`| HTTP request path |
|`pathParameters`| List of parameters which will be mapped to command flags, and used to substitute any placeholders in the `.path` property |

Notice the `path` is using a variable, `{id}`. This means that it requires a parameter to be defined elsewhere so that the REST API request gets can replace the `{id}` placeholder with the asset id given by the user. Since it the `{id}` is defined in the path, it means that we need to define a parameter under the `pathParameters`. Each of these parameters will be exposed to the user via flags of the command line.

Now let's check if the command is doing what it should. Running the command with the --dry argument is recommended until you're sure you did everything correctly.


<CodeExample>

```bash
c8y organizer assets get --id 12345 --dry
```

</CodeExample>

```bash title="Output"
What If: Sending [GET] request to [https://{host}/service/organizer/12345]

### GET /service/organizer/12345
```

Though let's say that our fictitious API also supports a query parameter to determine if the user would like to include extra details about the device, then we can modify the above snippet to also provide a `queryParameters` section which accepts a list of parameter (exactly like the `pathParameters`) however the parameters are mapped to query parameters instead of to the path.

Below adds the `detailed` parameter which is a boolean. The boolean type is mapped to a flag which does not accept an argument, e.g. `--detailed`. If `--detailed` is not present, then the query parameter is not added to the outgoing request.

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/feat/extensions-manager/tools/schema/extensionCommands.json
commands:
  - name: get
    description: Get asset
    method: GET
    path: service/organizer/{id}
    pathParameters:
      - name: id    # <= This should have the same name as placeholder in the .path field!
        type: string
        description: Asset
    
    queryParameters:
      - name: detailed
        type: boolean
        description: Include detailed values
```

Below show the new `detailed` parameter in action:

<CodeExample>

```bash
c8y organizer assets get --id 12345 --detailed --dry
```

</CodeExample>

```bash title="Output"
What If: Sending [GET] request to [https://{host}/service/organizer/12345?detailed=true]

### GET /service/organizer/12345?detailed=true
```

:::note
There are a lot of different types that you can use when building your commands (not just `string`, `boolean` etc.), so there is likely to be one that meets your need.
:::


### Step 4: Adding tab completion

Tab completion is a very useful feature which saves the user looking up things themselves as they can just press `<TAB><TAB>` and select an option from the response.

Some parameter types (such as `[]device` and `[]application`) include built-in tab completion and named lookups, however if you don't find any types that meet your exact need then you can use the external tab completion option. The external tab completion mechanism allows you to execute another `c8y` command, or a shell of your choosing, to provide the completion values that should be displayed to the user.

Below shows an example of an external completion which uses the `c8y devices list` to provide the device names (with the device id being shown in the option's description for more context).

```yaml
commands:
  - name: get
    description: Get asset
    method: GET
    path: service/organizer/{id}
    pathParameters:
      - name: id
        type: string
        description: Asset
        completion:   # <= Provide tab completion via another c8y command
          type: external
          command:
            - c8y
            - devices
            - list
            - --query
            - name eq '%s'
            - --select=name,id
```

The following shows the example tab completion output which is now presented to the user.

<CodeExample>

```bash
c8y organizer assets get --id <TAB><TAB>
```

</CodeExample>

```bash title="Output"
linux_001  --  | id: 441672938
linux_002  --  | id: 401669543
```

:::note
The `--select <cols>` flag used in the completion command is used to control which data . The first value is the value which will be returned to the user, and the additional columns are used to provide additional context to the user to help with the selection.
:::

:::tip
Any command including extension commands can be used to provide tab completion options. This enable maximum flexibility when creating your extension.
:::


### Step 5: Adding named lookups

A named lookups are similar to the external completion functions, as they allow you to define a function which takes in the user's selection and returns the actual value that should be used in the request. For example to support named lookups in the `--id` flag, the command needs to define a function to convert the name to an id.

In most cases you can re-use the completion command and just change the `--select` to return the id. The named lookup command can be provided by the `lookup` property of a parameter. Below shows an example which re-uses the `c8y devices list` command to find any matching. Take special note of the usage of `%s` without the command. The `%s` is substituted with the current value provided by the user.

```yaml
commands:
  - name: get
    description: Get asset
    method: GET
    path: service/organizer/{id}
    pathParameters:
      - name: id
        type: string
        description: Asset
        completion:
          type: external
          command:
            - c8y
            - devices
            - list
            - --query
            - name eq '%s'
            - --select=name,id
        lookup:     # <= Lookup a value by name
          type: external
          command:
            - c8y
            - devices
            - list
            - --query
            - name eq '%s*' and has(c8y_IsLinux)
            - --select=id
```

Below shows the lookup by name in action. The `linux_001` should be replaced by the actual id of the value.

<CodeExample>

```bash
c8y organizer assets get --id linux_001 --dry
```

</CodeExample>

```bash title="Output"
What If: Sending [GET] request to [https://{host}/service/organizer/441672938]

### GET /service/organizer/441672938
```

:::note
If the lookup command returns multiple matches, the first result will be used.
:::


### Step 6: Activate pipeline functionality

By default a command is not automatically pipeline enabled, you will have to mark a specific parameter as default flag. Any parameter/flag is allowed to accept pipeline input.

Without pipeline support trying to use the command with piped input will result in errors. Below shows an example of such errors:

<CodeExample>

```sh
# This won't work
echo 12345 | c8y organizer assets get --dry
```

</CodeExample>

```bash title="Output"
2023-05-07T16:56:35.442+0200    ERROR   commandError: missing required parameters. [id]
```

We can fix the situation by simply adding the `pipeline: true` option to the parameter. We'll pick the `id` parameter to be the default parameter where the piped input is mapped to as it will make the command the most useful.

```yaml
commands:
  - name: get
    description: Get asset
    method: GET
    path: service/organizer/{id}
    pathParameters:
      - name: id
        type: string
        description: Asset
        pipeline: true      # <= Activate pipeline mapping for the --id flag
        pipelineAliases:    # <= Optional: Additional properties to look for the --id value when json is being piped
          - deviceId
          - source.id
          - id
        completion:
          type: external
          command:
            - c8y
            - devices
            - list
            - --query
            - name eq '%s*' and has(c8y_IsLinux)
            - --select=name,type,id
        lookup:
          type: external
          command:
            - c8y
            - devices
            - list
            - --query
            - (name eq '%s*') and has(c8y_IsLinux)
            - --select=id
    queryParameters:
      - name: detailed
        type: boolean
        description: Include detailed values
```

After the `--id` parameter has been marked as accepting pipeline, then any input data is now magically mapped to the `id` flag, take special note of the HTTP path variable.

<CodeExample>

```sh
echo 12345 | c8y organizer assets get --dry
```

</CodeExample>

```bash title="Output"
What If: Sending [GET] request to [https://{host}/service/organizer/12345]

### GET /service/organizer/12345
```

:::note
If you forget to mark a flag with the `pipeline: true` option, then the user can still manually map pipeline input to a specific flag using the `-` value.

```sh
# --id will get the value from the piped input
echo 12345 | c8y organizer assets get --id - --dry
```
:::


### Step 7: Add remaining commands and use YAML anchors

YAML anchors can be used to minimize the amount of copy/pasting requires when creating a spec.

For example, we want to support external completion and named lookups on the `id` parameter, however it is used in multiple commands (e.g. `get`, `update` and `delete`). To prevent duplication we can create a custom type on the root level, give it an alias `type-asset` and then reference it later on.

To start off, let's create a re-usable snippet to contain the configuration for the `id` parameter and called it `x-type-asset`. It will be defined on the root level of the YAML document, and we'll assign it an anchor called `type-asset`. Below shows the snippet:


```yaml
# Use can use yaml anchors to reduce the amount of boilerplate
x-type-asset: &type-asset
  type: string
  description: Device. It support a custom completion/lookup using other c8y commands
  pipeline: true
  completion:
    type: external
    command:
      - c8y
      - devices
      - list
      - --query
      - "name eq '%s*'"
      - --select=name
  lookup:
    type: external
    command:
      - c8y
      - devices
      - list
      - --query
      - "name eq '%s*'"
      - --select=id
```

The `type-asset` anchor can then be reused through the YAML specification where applicable.

Below shows the final API specification after all the commands have been added to it and the `id` parameters reference the `type-asset` anchor (using the slightly obscure but useful YAML syntax `<<: *type-asset`):

```yaml title="file: api/assets.yaml"
# yaml-language-server: $schema=https://raw.githubusercontent.com/reubenmiller/go-c8y-cli/feat/extensions-manager/tools/schema/extensionCommands.json
---
group:
  name: assets
  description: Manage assets
  descriptionLong: |
    More descriptive block which can even list include example of how to use different commands together.
    c8y organizer devices list | c8y c8y organizer update --name "My name"

# Use YAML anchor which can be referenced by the parameters
x-type-asset: &type-asset
  type: string
  description: Asset
  pipeline: true
  pipelineAliases:
    - id
    - deviceId
    - source.id
  completion:
    type: external
    command:
      - c8y
      - devices
      - list
      - --query
      - name eq '%s*' and has(c8y_IsLinux)
      - --select=name,type,id
  lookup:
    type: external
    options:
      idPattern: '^[0-9]+$'
    command:
      - c8y
      - devices
      - list
      - --query
      - (name eq '%s*') and has(c8y_IsLinux)
      - --select=id

commands:
  - name: list
    description: Get asset collection
    method: GET
    path: service/organizer

  - name: get
    description: Get asset
    method: GET
    path: service/organizer/{id}
    pathParameters:
      - name: id
        <<: *type-asset
    queryParameters:
      - name: detailed
        type: boolean
        description: Include detailed values

  - name: update
    description: Update asset
    method: PUT
    path: service/organizer/{id}
    pathParameters:
      - name: id
        <<: *type-asset

  - name: delete
    description: Delete asset
    method: DELETE
    path: service/organizer/{id}
    pathParameters:
      - name: id
        <<: *type-asset
```
