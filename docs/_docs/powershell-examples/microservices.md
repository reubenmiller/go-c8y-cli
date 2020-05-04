---
layout: default
category: Examples - Powershell
title: Microservices
---

## Get

### Get a microservice by name

```powershell
Get-Microservice -Id helloworld
```

**Response**

```plaintext
id   name       key                         type         version availability requiredRoles
--   ----       ---                         ----         ------- ------------ -------------
9994 helloworld helloworld-microservice-key MICROSERVICE         MARKET       {}
```

### Get a list of microservices

List microservices being hosted in the platform

```powershell
Get-MicroserviceCollection
```

**Response**

```plaintext
id   name               key                  type         version         availability requiredRoles
--   ----               ---                  ----         -------         ------------ -------------
1001 sms-gateway        sms-gateway-key      MICROSERVICE 1005.6.1        MARKET       {ROLE_INVENTORY_READ, ROLE_INVENTORY_ADMIN, ROLE_IDENTITY_READ, ROLE_IDENTITY_ADMIN…}
11   smartrule          smartrule-key        MICROSERVICE 1005.6.1        MARKET       {ROLE_INVENTORY_READ, ROLE_INVENTORY_CREATE, ROLE_INVENTORY_ADMIN, ROLE_CEP_MANAGEMENT_READ…}
12   device-simulator   device-simulator-key MICROSERVICE 1005.6.1        MARKET       {ROLE_INVENTORY_READ, ROLE_INVENTORY_ADMIN, ROLE_INVENTORY_CREATE, ROLE_MEASUREMENT_READ…}
19   jwireless          jwireless-key        MICROSERVICE 1005.6.1        MARKET       {}
3147 apama-ctrl-starter apama-ctrl-starter   MICROSERVICE 10.5.0.3.363871 MARKET       {ROLE_APPLICATION_MANAGEMENT_READ, ROLE_APPLICATION_MANAGEMENT_ADMIN, ROLE_INVENTORY_READ, ROLE_INVENTORY_ADMIN…}
```

### Get a list of microservices with names starting with citest*

```powershell
Get-MicroserviceCollection -PageSize 100 |? name -like "citest*"
```

## Create

### Create a new microservice

The following command will create a new microservice, upload it's binary, and also subscribe to it on the current tenant:

```powershell
New-Microservice -File helloworld.zip
```

If you don't want to subscribe to the microservice immediately then use the `-SkipSubscription` option:

```powershell
# Create microservice but don't subscribe to it
New-Microservice -File helloworld.zip -SkipSubscription

# Enable/Subscribe to the microservice when you're ready
Enable-Microservice -Id helloworld
```

## Update

### Update the availability of the microservice to MARKET

```powershell
Update-Microservice -Id helloworld -Availability MARKET
```

**Response**

```plaintext
id   name       key                         type         version availability requiredRoles
--   ----       ---                         ----         ------- ------------ -------------
9994 helloworld helloworld-microservice-key MICROSERVICE         MARKET       {}
```

## Adding custom data to the application

```powershell
Update-Microservice -Id helloworld -Data @{ c8y_application_details = @{ branch = "master" } }
```

**Response**

```plaintext
Confirm
Are you sure you want to perform this action?
Performing the operation "Update microservice [helloworld (9994)]" on target "goc8yci01".
[Y] Yes  [A] Yes to All  [N] No  [L] No to All  [S] Suspend  [?] Help (default is "Y"):

id   name        key                           type         version  availability requiredRoles
--   ----        ---                           ----         -------  ------------ -------------
9994 helloworld  helloworld-microservice-key   MICROSERVICE          MARKET       {}
```

The full response can be printed to the console by piping the results to the `tojson` cmdlet.

```powershell
Get-Microservice -Id helloworld | tojson
```

**Response**

```json
{
  "owner": {
    "self": "https://goc8yci01.eu-latest.cumulocity.com/tenant/tenants/goc8yci01",
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
  "self": "https://goc8yci01.eu-latest.cumulocity.com/application/applications/9994",
  "id": "9994",
  "key": "helloworld-microservice-key",
  "c8y_application_details": {
    "branch": "master"
  }
}
```

## Delete/Remove

### Remove microservice

```powershell
Get-Microservice -Id helloworld
```

**Response**

```plaintext
id   name       key                         type         version availability requiredRoles
--   ----       ---                         ----         ------- ------------ -------------
9994 helloworld helloworld-microservice-key MICROSERVICE         MARKET       {}
```

### Remove microservices with starting with "citest"

```powershell
Get-MicroserviceCollection -PageSize 100 |? name -like "citest*" |
    Remove-Microservice
```

**Response**

```plaintext
Confirm
Are you sure you want to perform this action?
Performing the operation "Remove microservice [citestr16cj (8977)]" on target "goc8yci01".
[Y] Yes  [A] Yes to All  [N] No  [L] No to All  [S] Suspend  [?] Help (default is "Y"):
```

In PowerShell `?` is an alias for the `Where-Object` cmdlet, so the above example is equivalent to:

```powershell
Get-MicroserviceCollection -PageSize 100 |
    Where-Object { $_.name -like "citest*" } |
    Remove-Microservice
```

## Enable/Disable a microservice

Enabling a microservice can be done using:

```powershell
Enable-Microservice -Id helloworld
```

Once the microservice has started up (this can take a few minutes), then any endpoints made available by it, then it can be reached using the following:

```powershell
Invoke-CumulocityRequest -Uri "/service/helloworld/health"
```

To disable/unsubscribe a microservice from the current tenant use the following:

```powershell
Disable-Microservice -Id helloworld
```

## Advanced use cases

### Create a new microservice that will be hosted outside of Cumulocity (in private docker/kubernetes host)

```powershell
New-Microservice -File helloworld.zip -SkipUpload
```

The `-SkipUpload` will not upload the zip, however it will still parse the cumulocity.json manifest file which is used to update the microservice's required roles.

Then the microservice's bootstrap credentials can be retrieved using:

```powershell
Get-MicroserviceBootstrapUser -Id helloworld
```

**Response**

```plaintext
tenant         name                             password
------         ----                             --------
myTenant       servicebootstrap_helloworld      35aP2moL39zfe8PDo0OPH2D63kYhlqOG
```
