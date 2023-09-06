---
title: Events
---

import CodeExample from '@site/src/components/CodeExample';

## Download events binaries

You can download binaries (attachments) from a list of events by first listing the events, then piping the results to the download event binary commands. The binaries are downloaded to a file using the `outputFileRaw` flag which makes use of some variables to ensure each binary has a different name (based on the information stored in the event).

The following example assumes that all events with the type `c8y-configuration-plugin` have an attachment (otherwise an error will be printed on the console).

<CodeExample>

```bash
c8y events list --dateFrom -1d --type c8y-configuration-plugin \
| c8y events downloadBinary --noProgress --outputFileRaw "{filename}.txt"
```

</CodeExample>

After the command has finished you can view the saved files in the current directory.

```bash title="List files"
$ ls
3416200.txt	3425288.txt	3425295.txt	3426750.txt	3426768.txt	3427498.txt
```

If you don't know which events have binaries attached to them (and you don't care), then you can tell the command to silently ignore the "Not Found" (e.g. HTTP status code 404) errors.

It is still advised to use some sort of filter in the event list so that you don't return too many events as the command will still attempt to download an attachment regardless if it exists or not.

<CodeExample>

```bash
c8y events list --dateFrom -1d \
| c8y events downloadBinary --noProgress --outputFileRaw "{filename}.txt" --silentStatusCodes 404
```

</CodeExample>

:::info
The download binary command will still output the event binary's content to standard output (e.g. the console). If you are downloading binaries, then this can be very disruptive, therefore you can redirect the standard output to `/dev/null` using the following shell convention.

```sh
c8y events list --dateFrom -1d --type c8y-configuration-plugin \
| c8y events downloadBinary --noProgress --outputFileRaw "{filename}.txt" >/dev/null
```

:::

## Get counts of events in one hour intervals for the last 12 hours

:::caution
The shell example requires [gnu parallel](https://www.gnu.org/software/parallel/) to be installed
:::


<CodeExample>

```bash
# Required: gnu parallel
seq 0 11 | parallel -j1 --env --tags \
    c8y events list \
        --type device_connected \
        --dateFrom "-{}h" \
        --dateTo \"'-$(( {} - 1 ))h'\" \
        --withTotalPages \
        --pageSize 1 \
        --select self,statistics.totalPages \
        --output csv
```

```powershell
0..11 | Foreach-Object {
        Get-EventCollection `
            -Type device_connected `
            -DateFrom "-${PSItem}h" `
            -DateTo "-$( $PSItem - 2 )h" `
            -WithTotalPages `
            -PageSize 1 `
            -Select self,statistics.totalPages `
            -Output csv -Dry -DryFormat markdown
    }
```

</CodeExample>


```sh title="Output"
https://{tenant}/event/events?dateTo=2021-04-21T19:42:50.367513487%2B02:00&pageSize=1&dateFrom=2021-04-21T17:42:50.367508287%2B02:00&type=device_connected&currentPage=1&withTotalPages=true,0
https://{tenant}/event/events?dateTo=2021-04-21T17:42:50.629192387%2B02:00&pageSize=1&dateFrom=2021-04-21T15:42:50.629187187%2B02:00&type=device_connected&currentPage=1&withTotalPages=true,0
https://{tenant}/event/events?dateTo=2021-04-21T15:42:50.888966887%2B02:00&pageSize=1&dateFrom=2021-04-21T13:42:50.888961787%2B02:00&type=device_connected&currentPage=1&withTotalPages=true,0
https://{tenant}/event/events?dateTo=2021-04-21T13:42:51.100022987%2B02:00&pageSize=1&dateFrom=2021-04-21T11:42:51.100018187%2B02:00&type=device_connected&currentPage=1&withTotalPages=true,0
https://{tenant}/event/events?dateTo=2021-04-21T11:42:51.267113687%2B02:00&pageSize=1&dateFrom=2021-04-21T09:42:51.267106787%2B02:00&type=device_connected&currentPage=1&withTotalPages=true,0
https://{tenant}/event/events?dateTo=2021-04-21T09:42:51.453355987%2B02:00&pageSize=1&dateFrom=2021-04-21T07:42:51.453350587%2B02:00&type=device_connected&currentPage=1&withTotalPages=true,0
```

### Explanation (shell)

`seq` can is used to here to create to generate the relative `dateFrom` parameters. `seq 2 2 12` means to create a sequence of number start from 2 and stop at 12, but go in increments of 2. So running just the `seq` command gives the following output.

```sh
2
4
6
8
10
12
```

The output of the `seq` command is then piped into another useful utility `parallel` which takes input from standard input and passes it to the given command using the `--tags` parameter. Each piped input line is available for use by referencing the string `{}`. In this case `{}` is used for setting the `dateFrom` and `dateTo`.

`dateTo` is a little more complicated as it needs to do a small addition before it can be used as a relative date. `$(( <line_value> + 2 ))` is sh/bash/zsh syntax for performing arithmetic.

Technically `xargs` could also be used, however it is much more restrictive when it comes to handling special characters etc.

Below how the command will be executed by `parallel`. The additional parameters i.e. `withTotalPages` have been left out so it is easier to read.

```sh
c8y events list --type device_connected --dateFrom "-2h" --dateTo "-0h" 
c8y events list --type device_connected --dateFrom "-4h" --dateTo "-2h" 
c8y events list --type device_connected --dateFrom "-6h" --dateTo "-4h" 
c8y events list --type device_connected --dateFrom "-8h" --dateTo "-6h" 
c8y events list --type device_connected --dateFrom "-10h" --dateTo "-8h" 
c8y events list --type device_connected --dateFrom "-12h" --dateTo "-10h" 
```
