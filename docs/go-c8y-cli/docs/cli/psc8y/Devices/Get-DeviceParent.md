---
category: Devices
external help file: PSc8y-help.xml
id: Get-DeviceParent
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Devices/get-deviceparent
title: Get-DeviceParent
---



## SYNOPSIS
Get device parent references for a device

## SYNTAX

### ByLevel (Default)
```
Get-DeviceParent
	[[-Device] <Object[]>]
	[[-Level] <Int32>]
	[<CommonParameters>]
```

### Root
```
Get-DeviceParent
	[[-Device] <Object[]>]
	[-RootParent]
	[<CommonParameters>]
```

### All
```
Get-DeviceParent
	[[-Device] <Object[]>]
	[-All]
	[<CommonParameters>]
```

## DESCRIPTION
Get the parent of a device by using the references stored in the device managed object.

## EXAMPLES

### EXAMPLE 1
```
Get-DeviceParent device0*
```

Get the direct (immediate) parent of the given device

### EXAMPLE 2
```
Get-DeviceParent -All
```

Return an array of parent devices where the first element in the array is the root device, and the last is the direct parent of the given device.

### EXAMPLE 3
```
Get-DeviceParent -RootParent
```

Returns the root parent.
In most cases this will be the agent

## PARAMETERS

### -Device
Device id, name or object.
Wildcards accepted

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Level
Level to navigate backward from the given device to its parent/s
1 = direct parent
2 = parent of its parent
If the Level is too large, then the root parent will be returned

```yaml
Type: Int32
Parameter Sets: ByLevel
Aliases:

Required: False
Position: 2
Default value: 1
Accept pipeline input: False
Accept wildcard characters: False
```

### -RootParent
Return the top level / root parent

```yaml
Type: SwitchParameter
Parameter Sets: Root
Aliases:

Required: False
Position: Named
Default value: False
Accept pipeline input: False
Accept wildcard characters: False
```

### -All
Return a list of all parent devices

```yaml
Type: SwitchParameter
Parameter Sets: All
Aliases:

Required: False
Position: Named
Default value: False
Accept pipeline input: False
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

## NOTES

## RELATED LINKS
