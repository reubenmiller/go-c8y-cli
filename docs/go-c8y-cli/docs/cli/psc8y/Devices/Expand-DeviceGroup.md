---
category: Devices
external help file: PSc8y-help.xml
id: Expand-DeviceGroup
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Devices/expand-devicegroup
title: Expand-DeviceGroup
---



## SYNOPSIS
Expand a list of device groups

## SYNTAX

```
Expand-DeviceGroup
	[-InputObject] <Object[]>
	[<CommonParameters>]
```

## DESCRIPTION
Expand a list of device groups replacing any ids or names with the actual user object.

## EXAMPLES

### EXAMPLE 1
```
Expand-DeviceGroup "myGroup"
```

Retrieve the user objects by name or id

### EXAMPLE 2
```
Get-DeviceGroupCollection *test* | Expand-DeviceGroup
```

Get all the device groups (with "test" in their name).
Note the Expand cmdlet won't do much here except for returning the input objects.

## PARAMETERS

### -InputObject
List of ids, names or user objects

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
If the given object is already an user object, then it is added with no additional lookup

## RELATED LINKS
