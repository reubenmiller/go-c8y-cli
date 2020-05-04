---
layout: default
category: Examples - Powershell
title: Inventory
---

## Create

##### Create a new group, and device and assign the device to the group

```powershell
$Group = New-ManagedObject -Name "Reuben" -Data @{ c8y_IsDeviceGroup = @{} }
$Device = New-ManagedObject -Name "Device01" -Data @{ c8y_IsDevice = @{} }

New-ChildAssetReference -Group $Group -NewChildDevice $Device
```

## Update

##### Find a managed object by name, and then update it's name

```powershell
Find-ManagedObjectCollection -Query "name eq 'Reuben_u3migc60sn'" | Update-ManagedObject -Name "Reuben"
```

##### Adding update firmware and software capabilities (using supported operations)

```powershell
fmo -Query "name eq 'rmi_device01'" | Update-ManagedObject -Data @{ c8y_SupportedOperations = @("c8y_Firmware", "c8y_SoftwareList") }
```


##### Remove device fragment

```powershell
Find-ManagedObjectCollection -Query "name eq 'Reuben'" | Update-ManagedObject -Data @{ c8y_IsDevice = $null }
```



## Child Assets and Devices

##### Get a list of child assets of a group

```powershell
Get-ChildAssetCollection -Group Reuben
```
