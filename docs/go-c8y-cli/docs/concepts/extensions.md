---
title: Extensions
---

import CodeExample from '@site/src/components/CodeExample';

## Overview

Extensions allow you to customize go-c8y-cli to optimize you and your team's workflows. You can customize how your data is displayed and provide custom commands to simplify repetitive tasks.

Extensions utilize already existing features but package them as a git repository so they are easy to install and share. This makes it easy for a team to collaborate to add new commands for custom microservices, or just add custom columns to the device view so you can display new custom managed object fragments by default.

By default extensions are accessible across all sessions. Though for scenarios where you would like to limit an extension to a single session or a group of sessions you can change the folder where the extensions are stored based on customer or some other task orientated grouping.

## What is an extension?

An extension is a git repository which contains multiple files/folders which control what the extension has to offer. The extension creator is free to pick which elements should be included. By default all extensions start with `c8y-`. This is to make public extension easier to discover, however the `c8y-` prefix will be ignore when using the extension (to avoid unnecessary typing).

The different elements of an extension and where they are stored within the extension repository are listed below:

| Type | Path | Description |
|-----------|------|-------------|
| Aliases | extension.yaml | Convenience commands which are easily accessible under the root command. e.g. `c8y my-alias` |
| Commands | `commands/` | More complex commands which can be written in language  (e.g. bash, python etc.). The commands can call other go-c8y-cli commands or any other tooling |
| Templates | `templates/` | Any go-c8y-cli templates that can be referenced by when using the `template` flag |
| Views | `views/` | Any go-c8y-cli view definitions that control which fragments are shown for which items, e.g. show custom fragments for specific device types etc. |

Below shows an example extension `c8y-myext` and a tree representation of the files associated with it.

```
c8y-myext/
│
├── extension.yaml
│
├── commands/
│   ├── do-something
│   └── mysubcmd/
│       ├── list
│       └── get
│
├── templates/
│   ├── devices.json
│   └── applications.json
│
└── views/
    ├── devices.json
    └── applications.json
```

## Extension contents and examples

### Aliases

Aliases are defined in the `extension.yaml` file on the root level of the repository. The user can define any number of aliases, however aliases do not have any hierarchy so they should be reserved for commands which are frequently used and are a "time saver".

The aliases should not clash with any existing commands. If the alias is too specific then it might be better to leave the alias out and allow the users to specify their own session-based aliases using `c8y alias set`.

Read the [Aliases concept](https://goc8ycli.netlify.app/docs/configuration/aliases/) page for more details about it.

Below is an example of an `extension.yaml` file which defines one alias called `mo`. `mo` pretty prints a managed object as json when given a managed object's id.

```yaml title="file: extension.yaml"
version: "v1"
aliases:
  - name: mo
    description: Pretty print inventory managed object
    command: |
      inventory get --id "$1" --output json --view off
    shell: false
```

Once the extension is installed, the above alias is accessible using:

<CodeExample>

```bash
c8y mo 12345
```

</CodeExample>


### Commands

An extension can include any number of commands. The structure of the commands is based on the folder structure, so you can group commands by placing them under the same sub folder. There is no limit to the number of sub folders, however you should keep it under 4-5 levels so it is not annoying for users to type.

Below shows some examples of commands provided by an extension called `c8y-myext` and how each command can be executed.

| Path | Command called via |
|-----------|------|
| `./commands/services/list` | `c8y myext services list` |
| `./commands/services/get` | `c8y myext services get` |
| `./commands/list` | `c8y myext list` |

The commands themselves can be written in any script-based language, e.g. `bash`, `python`, `ruby` etc., however they should include a [Shebang](https://en.wikipedia.org/wiki/Shebang_(Unix)) so that the correct shell interpreter can be called by the operating system.

Below is a simple example of a bash-based script.

```bash title="file: ./commands/list"
#!/usr/bin/env bash
set -e
echo "Hey look at my" >&2
```

:::note
Note: Make sure that your file is executable. On Unix, you can execute `chmod +x file_name` in the command line to make file_name executable. On Windows, you can run `git init -b main`, `git add file_name`, then `git update-index --chmod=+x file_name`.
:::

### Templates

An extension can provide templates which are accessible when using the `template` flag. Templates can be more useful than commands when using them together with the template `var("name")` syntax which enables the user to customize the template values via the `templateVars` flag (which is tab completed).

Information about what a template is and how to create on can be found in the [Templates concept](https://goc8ycli.netlify.app/docs/concepts/templates/) page.

Below shows an small example of a `jsonnet` template to create a custom operation which accepts one template variable called `action`.

```jsonnet title="file: ./templates/custom.operation.jsonnet"
{
    "description": "Execute custom operation",
    "parameters": {
        "action": var("action", "init 6"),
    },
}
```

The template is accessible via the `template` flag and the template is prefixed with `<EXTENSION_NAME>::` (without the `c8y-` prefix).

<CodeExample>

```bash
c8y operations create --device 12345 --template myext::custom.operation.jsonnet --templateVars action="do_something_less_destructive"
```

</CodeExample>


### Views

Views allow you to custom what fragments are displayed by default for specific responses. A view definition has a selection criteria which controls when the view is activated and which columns are displayed on the console.

Checkout the [Views concept](https://goc8ycli.netlify.app/docs/concepts/views/) page for more details.

Like templates, extension views are also prefixed with `<EXTENSION_NAME>::` (without the `c8y-` prefix) to avoid name clashes amongst extensions and any other user-created views.

Views are generally automatically selected based on their activation criteria (and priority), but they can also be manually activated using the global `view` flag.

<CodeExample>

```bash
c8y operations create --device 12345 --view myext::mydevice
```

</CodeExample>

---

## How to use extensions

This section details how to interact with extensions.

### List already installed extensions

A list of the currently installed sessions can displayed using

<CodeExample>

```bash
c8y extensions list

# show details about the extensions
c8y extensions list --raw
```

</CodeExample>

### Installing a new extension

**Prerequisites**

Installing extensions requires the `git` command to be already installed. When an extension is installed from an external source, git is used to clone the repository to your file system. If you are cloning a private repository, then it is up to you to provide the necessary credentials when prompted. `go-c8y-cli` does not handle any of the repository authentication. If you are having problems then try cloning the repository manually.

If you do not have git on your machine, then you also install an extension from a local folder.

Extensions can be installed from the following locations:

* From a local folder
* From GitHub via `<owner>/<extension>`
* From another git repository hosting service


<CodeExample>

```bash
# From a local folder
c8y extensions install .

# From GitHub
c8y extensions install reubenmiller/c8y-myext

# From another git repository hosting service
c8y extensions install https://github.com/reubenmiller/c8y-myext.git
```

</CodeExample>

If the extension contains any commands, then they grouped under the extension name (without the `c8y-` prefix).

<CodeExample>

```sh
c8y myext list
```

</CodeExample>


:::caution
Never edit an extension directly from the `go-c8y-cli` extension folder as you will lose any unpublished changes if you call the `c8y extensions delete <name>` command!

Instead clone the extension manually and install it using the filesystem path to the cloned repo. This way `go-c8y-cli` will only create a symlink to the folder, so deleting the extension will only remove the symlink and not the original folder.
:::

### Deleting an extension

An extension can be removed by using the following command

<CodeExample>

```bash
c8y extensions delete my-extension
```

</CodeExample>

### Creating

To make it easier to create your own extensions, there is an in-built command which generates an extension with some examples. This is done using

<CodeExample>

```bash
c8y extensions create myext
```

</CodeExample>

Then follow the on-screen instructions for using it.

:::tip
Creating an extension requires `git` to already by installed.
:::

:::note
The extension will be automatically prefixed with `c8y-` so that all extensions follow the same convention making them easier to find on GitHub.
:::

### Updating an extension

Extensions can be updated using the following command. Any extensions which were installed from a local folder will be ignored.

<CodeExample>

```bash
# Update a single extension
c8y extensions update myext

# Update all extensions
c8y extensions update --all
```

</CodeExample>

Extensions are updated by `go-c8y-cli` by using standard git commands, `git fetch` and `git pull`.

## Advanced usage

This sections shows some more advanced use-cases. It is not intended for everyone, however you may find it useful in select scenarios.

### Set custom extensions location for a specific c8y session

If you want to isolate which extensions are used for a specific session or a group of sessions, then you can change the setting which controls where `go-c8y-cli` looks for extensions.

<CodeExample>

```bash
# 1. change to the session you want to change (if you have not already done this)
set-session

# 2. change the extension location
c8y settings update extensions.datadir "$HOME/my_customer/extensions/"

# 3. install the extension
c8y extensions install reubenmiller/c8y-example
```

</CodeExample>
