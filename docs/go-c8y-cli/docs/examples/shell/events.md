---
title: Events
---

## Get counts of events in two hour brackets for the last 12 hours

* Requires [gnu parallel](https://www.gnu.org/software/parallel/) to be installed

```sh
seq 2 2 12 | parallel -j1 --env --tags \
    c8y events list \
        --type device_connected \
        --dateFrom "-{}h" \
        --dateTo \"'-$(( {} + 2 ))h'\" \
        --withTotalPages \
        --pageSize 1 \
        --select self,statistics.totalPages \
        --output csv
```

*Output*

```sh
https://{tenant}/event/events?dateTo=2021-04-21T19:42:50.367513487%2B02:00&pageSize=1&dateFrom=2021-04-21T17:42:50.367508287%2B02:00&type=device_connected&currentPage=1&withTotalPages=true,0
https://{tenant}/event/events?dateTo=2021-04-21T17:42:50.629192387%2B02:00&pageSize=1&dateFrom=2021-04-21T15:42:50.629187187%2B02:00&type=device_connected&currentPage=1&withTotalPages=true,0
https://{tenant}/event/events?dateTo=2021-04-21T15:42:50.888966887%2B02:00&pageSize=1&dateFrom=2021-04-21T13:42:50.888961787%2B02:00&type=device_connected&currentPage=1&withTotalPages=true,0
https://{tenant}/event/events?dateTo=2021-04-21T13:42:51.100022987%2B02:00&pageSize=1&dateFrom=2021-04-21T11:42:51.100018187%2B02:00&type=device_connected&currentPage=1&withTotalPages=true,0
https://{tenant}/event/events?dateTo=2021-04-21T11:42:51.267113687%2B02:00&pageSize=1&dateFrom=2021-04-21T09:42:51.267106787%2B02:00&type=device_connected&currentPage=1&withTotalPages=true,0
https://{tenant}/event/events?dateTo=2021-04-21T09:42:51.453355987%2B02:00&pageSize=1&dateFrom=2021-04-21T07:42:51.453350587%2B02:00&type=device_connected&currentPage=1&withTotalPages=true,0
```

### Explanation

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

Below how the command will be executed by `parallel`. Note the additional parameters i.e. `withTotalPages` have been left out so it is easier to read.

```sh
c8y events list --type device_connected --dateFrom "-2h" --dateTo "-0h" 
c8y events list --type device_connected --dateFrom "-4h" --dateTo "-2h" 
c8y events list --type device_connected --dateFrom "-6h" --dateTo "-4h" 
c8y events list --type device_connected --dateFrom "-8h" --dateTo "-6h" 
c8y events list --type device_connected --dateFrom "-10h" --dateTo "-8h" 
c8y events list --type device_connected --dateFrom "-12h" --dateTo "-10h" 
```
