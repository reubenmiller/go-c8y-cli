---
layout: default
category: Concepts
title: Bash - Filtering
---

### Filtering (client side)

The response returned from the platform can also be filtered on the client side by using the `--filter` argument.

#### Filtering application with name that start with "co*"

```sh
applications list --pageSize 100 --filter "name like co*"
```

### Selecting properties

The `select` parameter was re-worked in v2.0.0, and information about how to use it has been moved to[shell select parameter](https://reubenmiller.github.io/go-c8y-cli/docs/concepts/Shell-Select Parameter/)

### Formatting data (removed in v2.0.0)

The `--format` parameter has been removed replaced by combining the use of `--select` and `--output csv`. Please read the section on outputing data as csv.
