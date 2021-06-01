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


## Examples

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
