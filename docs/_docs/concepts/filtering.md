---
layout: default
category: Concepts
title: Shell - Filtering
---

### Filtering (client side)

The response returned from the platform can also be filtered on the client side by using the `--filter` argument.

#### Filtering application with name that start with "co*"

```sh
applications list --pageSize 100 --filter "name like co*"
```
