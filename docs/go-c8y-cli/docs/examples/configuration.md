---
title: Configuration
---

import CodeExample from '@site/src/components/CodeExample';

## Replace domain used in urls

When uploading binaries to a configuration repository item, it can use the Cumulocity IoT url which includes the tenant name, for example a url like `https://t12345.cumulocity.com/inventory/binaries/11111`. Depending on your tenant setup, this URL might not be publicly reachable which might cause a problem to any agents/microservices want to download the binary behind the url.

Let's say that our Cumulocity IoT tenant is using a custom domain called `https://mycompany.iot.com`, however the configuration repository items are using urls the following the format; `https://t12345.cumulocity.com/inventory/binaries/11111`.

Replacing the domain can be done using the following chained command.

<CodeExample>

```shell
c8y configuration list --query "url eq '*https://t*/inventory/binaries/*'" \
| c8y inventory update \
    --template "{url: '$C8Y_HOST' + _.GetURLPath(input.value.url)}" \
    --dry
```

```powershell
c8y configuration list --query "url eq '*https://t*/inventory/binaries/*'" \
| c8y inventory update \
    --template "{url: '$env:C8Y_HOST' + _.GetURLPath(input.value.url)}" \
    --dry
```

```powershell
Get-ConfigurationCollection -Query "url eq '*https://t*/inventory/binaries/*'" `
| Update-ManagedObject `
    -Template "{url: '$env:C8Y_HOST' + _.GetURLPath(input.value.url)}" `
    -Dry
```

</CodeExample>


### Explanation

The solution involves two `c8y` commands:

1. Find the configuration repository items which have a url which includes the tenant internal path (e.g. `https://t12345.cumulocity.com/...`)
2. Update the `.url` field by replacing the existing host with the new one.

The first command finds the configuration repository items. It is important to use a query that only matches configuration items that are using the internal domain name. By doing this it ensures that the configuration item will not modified again if the command is run multiple times.

The second command uses a jsonnet template to do the domain name replacement. It uses the special `_.GetURLPath()` function provided by `go-c8y-cli` to make the manipulation of the URL much easier. The function returns the URL path without the domain/host. The function is applied to the configuration url by accessing the piped data (coming from the upstream configuration list command). The `input.data` contains the full representation of the current configuration item, so the url can be read by using using `input.data.url`. In addition the `C8Y_HOST` environment variable is referenced in the template (via shell expansion) so that you don't have to manually type the domain name making the command also portable to other tenants.

Before changing any data in the platform, it is best practice to validate the command first by following the points below:

* Set the `pageSize` to `1` on the first command to limit the number of items.
* Use the `dry` flag on the command which is modifying data. In this case it is the inventory update command.

We can check what the inventory update command will do (without sending any requests) for 1 configuration item using the following command:

<CodeExample>

```shell
c8y configuration list --query "url eq '*https://t*/inventory/binaries/*'" --pageSize 1 \
| c8y inventory update \
    --template "{url: '$C8Y_HOST' + _.GetURLPath(input.value.url)}" \
    --dry
```

```powershell
c8y configuration list --query "url eq '*https://t*/inventory/binaries/*'" --pageSize 1 \
| c8y inventory update \
    --template "{url: '$env:C8Y_HOST' + _.GetURLPath(input.value.url)}" \
    --dry
```

```powershell
Get-ConfigurationCollection -Query "url eq '*https://t*/inventory/binaries/*'" -PageSize 1 `
| Update-ManagedObject `
    -Template "{url: '$env:C8Y_HOST' + _.GetURLPath(input.value.url)}" `
    -Dry
```

</CodeExample>


````markdown title="Output"
### PUT /inventory/managedObjects/541184560

| header            | value
|-------------------|---------------------------
| Accept            | application/json
| Authorization     | Bearer {token}
| Content-Type      | application/json

#### Body

```json
{
  "url": "https://mycompany.iot.com/inventory/binaries/1178209"
}
```
````

The output shows it would send a `PUT` request to the managed object api, and it will only update the `.url` fragment, and the `.url` fragment includes the correct domain name. So it looks like the command is doing everything it should, now the pageSize can be increased (or you can use the `includeAll` flag instead), and the `dry` flag can be removed.

<CodeExample>

```shell
c8y configuration list --query "url eq '*https://t*/inventory/binaries/*'" --includeAll \
| c8y inventory update \
    --template "{url: '$C8Y_HOST' + _.GetURLPath(input.value.url)}"
```

```powershell
c8y configuration list --query "url eq '*https://t*/inventory/binaries/*'" --includeAll \
| c8y inventory update \
    --template "{url: '$env:C8Y_HOST' + _.GetURLPath(input.value.url)}"
```

```powershell
Get-ConfigurationCollection -Query "url eq '*https://t*/inventory/binaries/*'" -IncludeAll `
| Update-ManagedObject `
    -Template "{url: '$env:C8Y_HOST' + _.GetURLPath(input.value.url)}"
```

</CodeExample>
