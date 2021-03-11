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
| m | Get-Measurement |
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

---

### Bash / zsh

The following aliases are defined for both bash and zsh in the `c8y.plugin.sh` (for bash) and `c8y.plugin.zsh` (for zsh, oh-my-zsh).

| Alias | Command |
|-------|---------|
| alarm | `c8y alarms get --id` |
| alarms | `c8y alarms list` |
| app | `c8y applications get --id` |
| apps | `c8y applications list` |
| devices | `c8y devices list` |
| event | `c8y events get --id` |
| events | `c8y events list` |
| fmo | `c8y inventory find --query` |
| m | `c8y measurements get --id` |
| measurements | `c8y measurements list` |
| mo | `c8y inventory get --id` |
| op | `c8y operations get --id` |
| ops | `c8y operations list` |
| series | `c8y measurements getSeries` |
