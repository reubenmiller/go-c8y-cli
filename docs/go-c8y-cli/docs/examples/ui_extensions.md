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

But default the `cumulocity.json` manifest inside the zip file will be read to extract the default information, however you can manually override this information by providing the information on the command line.

For example, you can provide a manual extension name and version using:

<CodeExample transform="false">

```bash
c8y ui extensions create --file ./my_extension.zip --name custom_name --version 1.0.0 --tags latest
```

</CodeExample>

## List extensions

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