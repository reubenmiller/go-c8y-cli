---
layout: default
category: Configuration
title: Aliases
---

### Powershell

In order to make common commands easier/quicker to use, a few default aliases can be registered using:

```powershell
Register-Alias
```

These same aliases can be unregistered using:

```powershell
Unregister-Alias
```

Below is a list of the aliases:


| Alias | Command |
|-------|---------|
| alarm | Get-Alarm |
| alarms | Get-AlarmCollection |
| app | Get-Application |
| apps | Get-ApplicationCollection |
| devices | Get-DeviceCollection |
| event | Get-Event |
| events | Get-EventCollection |
| fmo | Find-ManagedObjectCollection |
| fromjson | ConvertFrom-Json |
| json | ConvertTo-Json |
| m | Get-Measurements |
| measurements | Get-MeasurementCollection |
| mo | Get-ManagedObject |
| op | Get-Operation |
| ops | Get-OperationCollection |
| rest | Invoke-RestRequest |
| series | Get-MeasurementSeries |
| session | Get-Session |
| tojson | ConvertTo-Json |

Custom Alias can still be registered by the user using the in-built Powershell command `Set-Alias`:

```powershell
Set-Alias -Name "series" -Value "Get-MeasurementSeries"
```

Or if you want to add create a simple wrapper where you can control some of it's arguments, you can create a new Function:

```powershell
# Create simple function wrapper around Get-ApplicationCollection
Function Get-MyApplications {
    [cmdletbinding()]
    Param()
    Get-ApplicationCollection -Type "MICROSERVICE" -PageSize 2000 |
        Where-Object { $_.name -like "myapp*" }
}

## Usage
Get-MyApplications
```
