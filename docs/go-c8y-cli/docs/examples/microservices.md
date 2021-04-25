---
title: Microservices
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Get

### Get a microservice by name

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y microservices get --id helloworld
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-Microservice -Id mytestapp
```

</TabItem>
</Tabs>


```plaintext title="Output"
| id         | name           | key            | type              | manifest.version | availability | resources  | owner.tenant.id | requiredRoles.0           |
|------------|----------------|----------------|-------------------|------------------|--------------|------------|-----------------|---------------------------|
| 14114      | helloworld     | helloworld     | MICROSERVICE      |                  | MARKET       |            | t12345         | ROLE_INVENTORY_ADMIN      |
```

### Get a list of microservices

List microservices being hosted in the platform

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y microservices list
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-MicroserviceCollection
```

</TabItem>
</Tabs>

```plaintext title="Output"
| id         | name       | key          | type              | manifest.version | availability | resources  | owner.tenant.id |
|------------|------------|--------------|-------------------|------------------|--------------|------------|-----------------|
| 12         | cep        | cep-key      | MICROSERVICE      | 1007.1.0        | MARKET        |            | management      |
| 14114      | helloworld | helloworld   | MICROSERVICE      |                 | MARKET        |            | t12345        |
| 18         | device-si… | device-simu… | MICROSERVICE      | 1007.1.0        | MARKET        |            | management      |
| 25         | report-ag… | report-agen… | MICROSERVICE      | 1007.1.0        | MARKET        |            | management      |
| 29         | smartrule  | smartrule-k… | MICROSERVICE      | 1007.1.0        | MARKET        |            | management      |
```

### Get a list of microservices with names starting with smart*

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y microservices list --pageSize 100 --filter "name like smart*"
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-MicroserviceCollection -PageSize 100 -Filter "name like smart*"
```

</TabItem>
</Tabs>


```text title="Output"
| id         | name           | key                | type              | manifest.version | availability | resources  | owner.tenant.id |
|------------|----------------|--------------------|-------------------|------------------|--------------|------------|-----------------|
| 29         | smartrule      | smartrule-key      | MICROSERVICE      | 1007.1.0        | MARKET       |            | management      |
```

## Create

### Create a new microservice

The following command will create a new microservice, upload it's binary, and also subscribe to it on the current tenant:

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y microservices create --file helloworld.zip
```

</TabItem>
<TabItem value="powershell">

```powershell
New-Microservice -File helloworld.zip
```

</TabItem>
</Tabs>


If you don't want to subscribe to the microservice immediately then use the `skipSubscription` parameter:

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
# Create microservice but don't subscribe to it
c8y microservices create --file helloworld.zip --skipSubscription

# Enable/Subscribe to the microservice when you're ready
c8y microservices enable --id helloworld
```

</TabItem>
<TabItem value="powershell">

```powershell
# Create microservice but don't subscribe to it
New-Microservice -File helloworld.zip -SkipSubscription

# Enable/Subscribe to the microservice when you're ready
Enable-Microservice -Id helloworld
```

</TabItem>
</Tabs>

## Update

### Update the availability of the microservice to MARKET

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y microservices update --id helloworld --availability MARKET
```

</TabItem>
<TabItem value="powershell">

```powershell
Update-Microservice -Id helloworld -Availability MARKET
```

</TabItem>
</Tabs>


```plaintext title="Output"
| id         | name       | key          | type              | manifest.version | availability | resources  | owner.tenant.id |
|------------|------------|--------------|-------------------|------------------|--------------|------------|-----------------|
| 14114      | mytestapp  | mytestapp    | MICROSERVICE      |                  | MARKET       |            | t12345          |
```

## Adding custom data to the application

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y microservices update --id helloworld --template "{c8y_application_details: { branch: 'master' }}"
```

</TabItem>
<TabItem value="powershell">

```powershell
Update-Microservice -Id helloworld -Template "{c8y_application_details: { branch: 'master' }}"
```

</TabItem>
</Tabs>

```plaintext title="Output"
| id         | name       | key          | type              | manifest.version | availability | resources  | owner.tenant.id |
|------------|------------|--------------|-------------------|------------------|--------------|------------|-----------------|
| 14114      | mytestapp  | mytestapp    | MICROSERVICE      |                  | MARKET       |            | t12345          |
```

The full response can be printed to the console by setting the `output` to `json` or using the `raw` option

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y microservices get --id helloworld --raw
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-Microservice -Id helloworld -Raw
```

</TabItem>
</Tabs>


```json title="Output"
{
  "owner": {
    "self": "https://mytenant.example.com/tenant/tenants/goc8yci01",
    "tenant": {
      "id": "goc8yci01"
    }
  },
  "requiredRoles": [],
  "manifest": {
    "noAppSwitcher": true,
    "settingsCategory": null
  },
  "roles": [],
  "contextPath": "helloworld",
  "availability": "MARKET",
  "type": "MICROSERVICE",
  "name": "helloworld",
  "self": "https://mytenant.example.com/application/applications/9994",
  "id": "9994",
  "key": "helloworld-microservice-key",
  "c8y_application_details": {
    "branch": "master"
  }
}
```

## Delete/Remove

### Remove microservice

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y microservices delete --id helloworld
```

</TabItem>
<TabItem value="powershell">

```powershell
Remove-Microservice -Id helloworld
```

</TabItem>
</Tabs>

```plaintext title="No Output"
✓ Deleted /application/applications/9994 => 204 No Content
```

### Remove microservices with starting with "citest"

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y microservices list --pageSize 100 --filter "name like citest*" |
  c8y microservices delete
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-MicroserviceCollection -PageSize 100 -Filter "name like citest*" | batch |
    Remove-Microservice
```

</TabItem>
</Tabs>


```plaintext title="Output"
✓ Deleted /application/applications/97388 => 204 No Content
```

## Enable/Disable a microservice

Enabling a microservice can be done using:

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y microservices enable --id helloworld
```

</TabItem>
<TabItem value="powershell">

```powershell
Enable-Microservice -Id helloworld
```

</TabItem>
</Tabs>


```text title="Output"
| application.id | application.name | application.type  | self                                                                             |
|----------------|------------------|-------------------|----------------------------------------------------------------------------------|
| 97388          | helloworld       | MICROSERVICE      | https://t12345.latest.stage.c8y.io/http://t12345.latest.stage.c8y.io/tenant… |
```

Once the microservice has started up (this can take a few minutes), then any endpoints made available by it, then it can be reached using the following:

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y api GET /service/helloworld/health --raw
```

</TabItem>
<TabItem value="powershell">

```powershell
Invoke-ClientRequest -Method "Get" -Uri "/service/helloworld/health" -Raw
```

</TabItem>
</Tabs>


```json title="Output"
{
  "status": "UP"
}
```

To disable/unsubscribe a microservice from the current tenant use the following:

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y microservices disable --id helloworld
```

</TabItem>
<TabItem value="powershell">

```powershell
Disable-Microservice -Id helloworld
```

</TabItem>
</Tabs>


```text title="Output"
✓ Deleted /tenant/tenants/t12345/applications/97388 => 204 No Content
```

## Advanced use cases

### Create a new microservice that will be hosted outside of Cumulocity (in private docker/kubernetes host)

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y microservices create --file helloworld.zip --skipUpload
```

</TabItem>
<TabItem value="powershell">

```powershell
New-Microservice -File helloworld.zip -SkipUpload
```

</TabItem>
</Tabs>


The `skipUpload` parameter tells the command to skip the binary upload, however it will still parse the cumulocity.json manifest file which is used to update the microservice's required roles.

Then the microservice's bootstrap credentials can be retrieved using:

<Tabs
  groupId="shell-types"
  defaultValue="bash"
  values={[
    { label: 'Shell', value: 'bash', },
    { label: 'PowerShell', value: 'powershell', },
  ]
}>
<TabItem value="bash">

```bash
c8y microservices getBootstrapUser --id helloworld --raw
```

</TabItem>
<TabItem value="powershell">

```powershell
Get-MicroserviceBootstrapUser -Id helloworld -Raw
```

</TabItem>
</Tabs>


```json title="Output"
{
  "name": "servicebootstrap_helloworld",
  "password": "1dkd8ajd8DJ8djd9sk)lpoyHGGOpai8s",
  "tenant": "t12345"
}
```
