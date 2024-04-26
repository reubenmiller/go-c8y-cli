---
title: UI Plugins
---

import CodeExample from '@site/src/components/CodeExample';
import Video from '@site/src/components/video';

## Basic

### Installing a new version of a plugin

The `c8y ui plugins create` command is a "smart" command which will perform all the required steps to create a new version of a plugin in Cumulocity IoT. The command will automatically check if an application placeholder needs to be created or not and add the binary (reading the version from the manifest inside the zip file). This takes allows users to focus on deploying new plugins easily.

<CodeExample transform="false">

```bash
c8y ui plugins create --file ./my_plugin.zip --tags latest
```

</CodeExample>

Alternatively, you can install a plugin directly from a URL (provided the URL does not require any authentication):

<CodeExample transform="false">

```bash
c8y ui plugins create --file "https://github.com/thin-edge/tedge-container-plugin/releases/download/1.2.3/tedge-container-plugin-ui_1.0.2.zip" --tags latest
```

</CodeExample>

But default the `cumulocity.json` manifest inside the zip file will be read to extract the default information, however you can manually override this information by providing the information on the command line.

For example, you can provide a manual plugin name and version using:

<CodeExample transform="false">

```bash
c8y ui plugins create --file ./my_plugin.zip --name custom_name --version 1.0.0 --tags latest
```

</CodeExample>

### List plugins

You can list local plugins in your tenant using:

<CodeExample transform="false">

```bash
c8y ui plugins list --pageSize 100
```

</CodeExample>

Or if you would like to view the shared plugins which are made available via the parent tenant you can use:

<CodeExample transform="false">

```bash
c8y ui plugins list --pageSize 100 --availability SHARED
```

</CodeExample>

### Update/Replace tags

You can replace the tags of an existing version using the following command:

For example the `latest` and `other` tags

<CodeExample transform="false">

```bash
c8y ui plugins versions update --plugin myplugin --version "1.0.0" --tags latest,other
```

</CodeExample>

### Delete a version

Deleting a version can either be done by referencing the tag or version.

#### Delete by tag

<CodeExample transform="false">

```bash
c8y ui plugins versions delete --plugin myplugin --tag other
```

</CodeExample>

#### Delete by version

<CodeExample transform="false">

```bash
c8y ui plugins versions delete --plugin myplugin --version "1.0.0"
```

</CodeExample>


## Advanced

### Only keep last N versions

You can easily run a command to regularly cleanup older versions of a plugin by combining the go-c8y-cli command with linux utilities, `sort` and `tail`.

Below will lists the versions of a given plugin only selecting the version (as text), then sort the list based on "human-version" sorting (e.g. 2.5.10 > 2.5.2), skips the first 2 versions, and pipes the remaining of the versions to the delete command.

<CodeExample transform="false">

```bash
c8y ui plugins versions list --plugin mycustom-ui-plugin --includeAll --select version -o csv \
| sort -rV \
| tail -n+3 \
| c8y ui plugins versions delete --plugin mycustom-ui-plugin --silentExit --silentStatusCodes 409
```

</CodeExample>

**Note** If the version that is tagged as *latest* is not included in the recent N versions, then you will end up with N+1 versions as Cumulocity IoT will prevent you from deleting the version marked as *latest*.

### Delete all versions by the latest

Cumulocity IoT will not allow you to delete a plugin which is tagged as "latest", therefore this makes it very easy to delete all of the versions by using the following 

<CodeExample transform="false">

```bash
c8y ui plugins versions list --plugin mycustom-ui-plugin \
| c8y ui plugins versions delete --plugin mycustom-ui-plugin
```

</CodeExample>

If you don't want the error message being displayed about trying to delete the version with the *latest* tag, then you can use the following:

<CodeExample transform="false">

```bash
c8y ui plugins versions list --plugin mycustom-ui-plugin \
| c8y ui plugins versions delete --plugin mycustom-ui-plugin --silentExit --silentStatusCodes 409
```

</CodeExample>

### List the plugins and check how many versions exist for each

Getting a list of plugins and how many versions exist for each plugin is useful to allow users to check if some versions need to be cleaned up or not.

<CodeExample transform="false">

```bash
c8y ui plugins list \
| c8y ui plugins versions list --pageSize 100 \
    --outputTemplate "input.value + {totalVersions: std.length(output)}" \
    --select id,name,totalVersions
```

</CodeExample>


```sh title="Output"
| id          | name                      | totalVersions |
|-------------|---------------------------|---------------|
| 102626      | plugin1                   | 1             |
| 69513       | lwm2m-ui-plugin           | 1             |
| 99228       | tedge-container-plugin-ui | 2             |
```

## Managing Application plugins

UI plugins can be installed into UI applications using the `c8y ui plugins install` command. This command can be used to install new plugins, upgrade existing plugins, and also replace a whole set of plugins installed in an application.

The follow sections detail the common use-cases.

### Install a plugin

A new plugin can be added to an existing UI application using the following command. The command will preserve any existing plugins, and it will only add a new one (or update the version if already installed).

```sh
c8y ui applications plugins install --application devicemanagement --plugin myplugin
```

By default the `latest` version is chosen, but a version or tag can be provided using the `@<tag|version>` syntax:

```sh
c8y ui applications plugins install --application devicemanagement --plugin myplugin@latest
```

Or install an explicit plugin version:

```sh
c8y ui applications plugins install --application devicemanagement --plugin myplugin@1.2.3
```

### Update all plugins to the latest versions

Existing UI plugins installed in an application can easily be updated using a single command.

```sh
c8y ui applications plugins update --application devicemanagement --all
```

You can update the plugins for all applications easily using:

```sh
c8y applications list --type HOSTED \
| c8y ui applications plugins update --all
```

### Replace all plugins with a new set of plugins

The list of UI plugins installed in an application can also be swapped out entirely by replacing all of the existing plugins with a new set of plugins.

```sh
c8y ui applications plugins replace --application devicemanagement --plugin myplugin --plugin another --plugin cloud-http-proxy
```

### Remove orphaned or revoked plugins from an application

Plugins and their versions can be removed at anytime, so sometimes an application will have references to plugins that don't exist, or a version that has been revoked/removed. These orphaned and revoked plugins can be removed using:

```sh
c8y ui applications plugins delete --application devicemanagement --invalid
```

Or if you need to cleanup the invalid plugins for a list of applications, then you can the pipeline.

```sh
c8y applications list --type HOSTED --name "devicemanagement*" \
| c8y ui applications plugins delete --application devicemanagement --invalid
```

### Remove all plugins from an application

Deleting all plugins from an existing application is done using:

```sh
c8y ui applications plugins delete --application devicemanagement --all
```
