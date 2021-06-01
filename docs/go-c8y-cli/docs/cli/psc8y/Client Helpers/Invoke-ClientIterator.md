---
category: Client Helpers
external help file: PSc8y-help.xml
id: Invoke-ClientIterator
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Client Helpers/invoke-clientiterator
title: Invoke-ClientIterator
---



## SYNOPSIS
Convert the input objects into a format that can be easily piped to the c8y binary directly

## SYNTAX

### string (Default)
```
Invoke-ClientIterator
	[-InputObject] <Object[]>
	[[-Format] <String>]
	[[-Repeat] <Int32>]
	[<CommonParameters>]
```

### json
```
Invoke-ClientIterator
	[-InputObject] <Object[]>
	[[-Repeat] <Int32>]
	[-AsJSON]
	[<CommonParameters>]
```

## DESCRIPTION
Calling the go c8y directly involves converting powershell objects either into json lines, or
just passing on the id.

The iterator can also also format the input data and fan it out (turning 1 input item into x items) by using the -Format and -Repeat parameters respectively.

## EXAMPLES

### EXAMPLE 1
```
Get-DeviceCollection | Invoke-ClientIterator | c8y devices update --data "mytype=myNewTypeValue"
Get a collection of devices and add a fragment "mytype: 'myNewTypeValue'" to each device.
```

### EXAMPLE 2
```
Get-Device myDeviceName | Invoke-ClientIterator -Repeat 5 | c8y measurements create --template example.jsonnet
```

Lookup a device by its name and then create 5 measurements using a jsonnet template

### EXAMPLE 3
```
@(1..20) | Invoke-ClientIterator "device" | c8y devices create --template example.jsonnet
```

Create 20 devices naming from "device0001" to "device0020" using a jsonnet template file.

### EXAMPLE 4
```
@(1..2) | Invoke-ClientIterator "device_{0}-{1}" -Repeat 2 | c8y devices create
```

Create 4 (Input count x Repeat) devices with the following names.

```powershell
device_1-0
device_1-1
device_2-0
device_2-1
```

### EXAMPLE 5
```
@(1..2) | Invoke-ClientIterator "device_{0}-{2}" -Repeat 2 | c8y devices create
```

Create 4 (Input count x Repeat) devices with the following names (using 1-indexed values when repeating)

```powershell
device_1-1
device_1-2
device_2-1
device_2-2
```

## PARAMETERS

### -InputObject
Input objects to be piped to native c8y binary

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: True
Position: 2
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Format
Format string to be applied to each value.
The format string is $Format -f $item
The value will be prefixed to the input objects by default.
However the format string
can be customized by using a powershell string format (i.e.
"{0:00}" )

Other format variables (additional )
"{0}" is the current input object (i.e.
{0:000} for 0 padded numbers)
"{1}" is the repeat counter from 0..Repeat-1
"{2}" is the repeat counter from 1..Repeat

```yaml
Type: String
Parameter Sets: string
Aliases:

Required: False
Position: 1
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Repeat
Repeat each input x times.
Useful when wanting to use the same item in multiple commands.
If a value less than 1 is provided, then it will be set to 1 automatically

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: 3
Default value: 0
Accept pipeline input: False
Accept wildcard characters: False
```

### -AsJSON
Convert the items to json lines

```yaml
Type: SwitchParameter
Parameter Sets: json
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
