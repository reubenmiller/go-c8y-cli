---
category: Devices
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Expand-Device
---

# Expand-Device

## SYNOPSIS
Expand a list of devices replacing any ids or names with the actual device object.

## SYNTAX

```
Expand-Device
	[-InputObject] <Object[]>
	[<CommonParameters>]
```

## DESCRIPTION
The list of devices will be expanded to include the full device representation by fetching
the data from Cumulocity.

## EXAMPLES

### EXAMPLE 1
```
Expand-Device "mydevice"
```

Retrieve the device objects by name or id

### EXAMPLE 2
```
Get-DeviceCollection *test* | Expand-Device
```

Get all the device object (with app in their name).
Note the Expand cmdlet won't do much here except for returning the input objects.

## PARAMETERS

### -InputObject
List of ids, names or device objects

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: True (ByValue)
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

## NOTES
If the given object is already an device object, then it is added with no additional lookup

## RELATED LINKS
