---
layout: default
category: Concepts
title: Filtering (client side)
---

import CodeExample from '@site/src/components/CodeExample';

## Supported operators

The response returned from the platform can also be filtered on the client side by using the `filter` argument.

The `filter` parameter uses a query language which supports the following operators:

|operator|description|usage|
|--------|-----------|-----|
|like|wildcard match (case insensitive)|`--filter "c8y_Hardware.serialNumber like *223*"`|
|notlike|inverted wildcard match (case insensitive)|`--filter "c8y_Hardware.serialNumber notlike *223*"`|
|match|regex match (case insensitive)|`--filter "c8y_Hardware.serialNumber match *223*"`|
|match|inverted regex match (case insensitive)|`--filter "c8y_Hardware.serialNumber notmatch *223*"`|
|eq|equals|`--filter "count eq 3"`|
|neq|not equals|`--filter "count neq 3"`|
|gt| greater than (numbers only)|`--filter "count gt 3"`|
|gte|greater than or equal to (numbers only)|`--filter "count ge 3"`|
|lt|less than (numbers only)|`--filter "count lt 3"`|
|lte|less than or equal to (numbers only)|`--filter "count le 3"`|
|leneq|equality match length of (string/array/map)|`--filter "name leneq 10"` or `--filter "childAdditions.references leneq 1"`|
|lenneq|inverted equality match length of (string/array/map)|`--filter "name lenneq 10"` or `--filter "childAdditions.references lenneq 1"`|
|lengt|match length greater than value of (string/array/map)|`--filter "name lengt 10"` or `--filter "childAdditions.references lengt 1"`|
|lengte|match length greater than or equal to value of (string/array/map)|`--filter "name lengte 10"` or `--filter "childAdditions.references lengte 1"`|
|lenlt|match length less than value of (string/array/map)|`--filter "name lenlt 10"` or `--filter "childAdditions.references lenlt 1"`|
|lenlte|match length less than or equal to value of (string/array/map)|`--filter "name lenlte 10"` or `--filter "childAdditions.references lenlte 1"`|
|datelt|match date less than (older) to value of (datetime|relative)|`--filter "creationTime datelt 2022-01-02T12:00"`|
|datelte (or 'olderthan')|match date less than (older) or equal to value of (datetime|relative)|`--filter "creationTime datelte 2022-01-02T12:00"`|
|dategt|match date greater than (newer) to value of (datetime|relative)|`--filter "creationTime dategt 2022-01-02T12:00"`|
|dategte (or 'newerthan')|match date greater than (newer) or equal to value of (datetime|relative)|`--filter "creationTime dategte 2022-01-02T12:00"`|

## Examples

:::caution
It is highly recommended to use server side filtering where you can (e.g. inventory query api). Server side filtering is much more efficient and it reduces the amount of data that the server needs to transfer to the client.

But if you have a very specific use-case which can't be solved via server API then this section is for you. Remember you an also use `jq` and `grep` if can't find the right client side filtering operator.
:::

### Filtering application with name that start with "co*"

<CodeExample>

```bash
c8y applications list --pageSize 100 --filter "name like co*"
```

</CodeExample>

```csv title="output"
| id         | name         | key                          | type        | availability |
|------------|--------------|------------------------------|-------------|--------------|
| 8          | cockpit      | cockpit-application-key      | HOSTED      | MARKET       |
```


### Filtering application with name that start with "co*"

<CodeExample>

```bash
c8y inventory list --pageSize 100 --filter "lastUpdated newerthan -10d"
```

</CodeExample>

```csv title="output"
| id         | name         | key                          | type        | availability |
|------------|--------------|------------------------------|-------------|--------------|
| 8          | cockpit      | cockpit-application-key      | HOSTED      | MARKET       |
```

### Filtering list of inventory (client side) by lastUpdated date between a given range

A date string can be used and it will be converted to your local timezone. i.e. "2022-01-18" -> "2022-01-18T00:00:00Z" when running on a machine using UTC time.

<CodeExample>

```bash
c8y inventory list \
    --pageSize 20 \
    --filter "lastUpdated newerthan 2022-01-18" \
    --filter "lastUpdated olderthan 2022-01-18 15:00" \
    --select id,lastUpdated
```

</CodeExample>

```csv title="output"
| id         | lastUpdated                   |
|------------|-------------------------------|
| 4207       | 2022-01-18T08:20:59.301Z      |
| 4209       | 2022-01-18T08:21:00.204Z      |
| 3211       | 2022-01-18T08:21:01.394Z      |
| 2702       | 2022-01-18T09:29:23.990Z      |
| 4115       | 2022-01-18T09:29:26.280Z      |
| 4116       | 2022-01-18T09:29:28.761Z      |
| 4150       | 2022-01-18T09:45:14.382Z      |
| 2721       | 2022-01-18T09:45:16.703Z      |
| 4153       | 2022-01-18T09:45:19.232Z      |
| 4190       | 2022-01-18T11:26:46.037Z      |
| 4751       | 2022-01-18T11:26:48.982Z      |
| 2739       | 2022-01-18T11:26:52.188Z      |
| 4767       | 2022-01-18T14:55:24.515Z      |
| 2758       | 2022-01-18T14:55:26.860Z      |
| 2759       | 2022-01-18T14:55:29.349Z      |
```

### Filtering list of inventory (client side) by creationTime date using relative timestamp

A date string can be used and it will be converted to your local timezone. i.e. "2022-01-18" -> "2022-01-18T00:00:00Z" when running on a machine using UTC time.

<CodeExample>

```bash
c8y inventory list \
    --pageSize 2000 \
    --filter "creationTime newerthan -2h" \
    --select id,creationTime
```

</CodeExample>

```csv title="output"
| id         | creationTime                  |
|------------|-------------------------------|
| 53758      | 2022-02-16T18:37:26.682Z      |
| 53791      | 2022-02-16T18:49:35.689Z      |
| 55624      | 2022-02-16T18:59:43.222Z      |
```
