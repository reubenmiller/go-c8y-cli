---
title: UI Extensions
---

import CodeExample from '@site/src/components/CodeExample';
import Video from '@site/src/components/video';

## Basic

### Installing a new version of an extension

The `c8y ui extensions create` command is a "smart" command which will perform all the required steps to create a new version of an extension in Cumulocity IoT. The command will automatically check if an application placeholder needs to be created or not and add the binary (reading the version from the manifest inside the zip file). This takes allows users to focus on deploying new extensions easily.

<CodeExample transform="false">

```bash
c8y ui extensions create --file ./my_extension.zip --tags latest
```

</CodeExample>

Alternatively, you can install an extension directly from a URL (provided the URL does not require any authentication):

<CodeExample transform="false">

```bash
c8y ui extensions create --file "https://github.com/thin-edge/tedge-container-plugin/releases/download/1.2.3/tedge-container-plugin-ui_1.0.2.zip" --tags latest
```

</CodeExample>

But default the `cumulocity.json` manifest inside the zip file will be read to extract the default information, however you can manually override this information by providing the information on the command line.

For example, you can provide a manual extension name and version using:

<CodeExample transform="false">

```bash
c8y ui extensions create --file ./my_extension.zip --name custom_name --version 1.0.0 --tags latest
```

</CodeExample>

### List extensions

You can list local extensions in your tenant using:

<CodeExample transform="false">

```bash
c8y ui extensions list --pageSize 100
```

</CodeExample>

Or if you would like to view the shared extensions which are made available via the parent tenant you can use:

<CodeExample transform="false">

```bash
c8y ui extensions list --pageSize 100 --availability SHARED
```

</CodeExample>

### Update/Replace tags

You can replace the tags of an existing version using the following command:

For example the `latest` and `other` tags

<CodeExample transform="false">

```bash
c8y ui extensions versions update --extension myext --version "1.0.0" --tags latest,other
```

</CodeExample>

### Delete a version

Deleting a version can either be done by referencing the tag or version.

#### Delete by tag

<CodeExample transform="false">

```bash
c8y ui extensions versions delete --extension myext --version "1.0.0"
```

</CodeExample>

#### Delete by version

<CodeExample transform="false">

```bash
c8y ui extensions versions delete --extension myext --tag other
```

</CodeExample>


## Advanced

### Only keep last N versions

You can easily run a command to regularly cleanup older versions of an extension by combining the go-c8y-cli command with linux utilities, `sort` and `tail`.

Below will lists the versions of a given extension only selecting the version (as text), then sort the list based on "human-version" sorting (e.g. 2.5.10 > 2.5.2), skips the first 2 versions, and pipes the remaining of the versions to the delete command.

<CodeExample transform="false">

```bash
c8y ui extensions versions list --extension mycustom-ui-ext --includeAll --select version -o csv \
| sort -rV \
| tail -n+3 \
| c8y ui extensions versions delete --extension mycustom-ui-ext --silentExit --silentStatusCodes 409
```

</CodeExample>

**Note** If the version that is tagged as *latest* is not included in the recent N versions, then you will end up with N+1 versions as Cumulocity IoT will prevent you from deleting the version marked as *latest*.

### Delete all versions by the latest

Cumulocity IoT will not allow you to delete an extension which is tagged as "latest", therefore this makes it very easy to delete all of the versions by using the following 

<CodeExample transform="false">

```bash
c8y ui extensions versions list --extension mycustom-ui-ext \
| c8y ui extensions versions delete --extension mycustom-ui-ext
```

</CodeExample>

If you don't want the error message being displayed about trying to delete the version with the *latest* tag, then you can use the following:

<CodeExample transform="false">

```bash
c8y ui extensions versions list --extension mycustom-ui-ext \
| c8y ui extensions versions delete --extension mycustom-ui-ext --silentExit --silentStatusCodes 409
```

</CodeExample>

### List the extensions and check how many versions exist for each

Getting a list of extensions and how many versions exist for each extension is useful to allow users to check if some versions need to be cleaned up or not.

<CodeExample transform="false">

```bash
c8y ui extensions list \
| c8y ui extensions versions list --pageSize 100 \
    --outputTemplate "input.value + {totalVersions: std.length(output)}" \
    --select id,name,totalVersions
```

</CodeExample>


```sh title="Output"
| id          | name                      | totalVersions |
|-------------|---------------------------|---------------|
| 102626      | ext1                      | 1             |
| 69513       | lwm2m-ui-plugin           | 1             |
| 99228       | tedge-container-plugin-ui | 2             |
```

## Managing Application extensions

UI extensions can be installed into UI applications using the `c8y ui extensions install` command. This command can be used to install new extensions, upgrade existing extensions, and also replace a whole set of extensions installed in an application.

The follow sections detail the common use-cases.

### Installing an extension

A new extension can be added to an existing UI application using the following command. The command will preserve any existing extensions, and it will only add a new one (or update the version if already installed).

```sh
c8y ui extensions install --application devicemanagement --extension myext
```

By default the `latest` version is chosen, but a version or tag can be provided using the `@<tag|version>` syntax:

```sh
c8y ui extensions install --application devicemanagement --extension myext@latest
```

Or install an explicit extension version:

```sh
c8y ui extensions install --application devicemanagement --extension myext@1.2.3
```

### Update all extensions to the latest versions

The existing UI extensions installed in an application can easily be updated using a single command.

```sh
c8y ui extensions install --application devicemanagement --update-versions
```

### Replace all extensions with a new set of extensions

The list of UI extensions installed in an application can also be swapped out entirely by replacing all of the existing extensions with a new set of extensions. The `--replace` flag will ensure that the existing set of UI extensions is ignored, and only the new extensions will be used.

```sh
c8y ui extensions install --application devicemanagement --replace --extension myext --extension another --extension cloud-http-proxy
```

### Remove all extensions from an application

Similar to the previous example, you can also remove all extensions by also using `--replace`, but not providing any extensions to install.

```sh
c8y ui extensions install --application devicemanagement --replace
```
