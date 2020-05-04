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

In cases where you don't want all of the properties being returned in the json object, then a list of property names can be given using the `--select` argument.

Nested properties are also supported.

#### Example

##### Only return the "id", "name" and "owner.tenant" properties for each application

```sh
c8y applications list --pageSize 2 --select "id,name,owner.tenant"
```

**Response**

```json
[
  {
    "id": "1",
    "name": "devicemanagement",
    "tenant": {
      "id": "management"
    }
  },
  {
    "id": "10003",
    "name": "citest1jgrg",
    "tenant": {
      "id": "goc8yci01"
    }
  }
]
```

### Formatting data

If you would only like to return a single value, then you can use the `--format` argument to set the property that you would like to be returned.

This is ideal for getting the ids 

#### Example

##### Get the application id by looking it up by its name

```sh
c8y applications get --id cockpit --format "id"
```

**Response**

```plaintext
7
```

**Note:**

The formatting argument only supports one value. If the commands returns more than 1 result, then only the first result will be used.
