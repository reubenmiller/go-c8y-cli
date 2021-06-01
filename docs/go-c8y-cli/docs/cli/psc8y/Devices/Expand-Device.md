---
category: Devices
external help file: PSc8y-help.xml
id: Expand-Device
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Devices/expand-device
title: Expand-Device
---



## SYNOPSIS
Expand a list of devices replacing any ids or names with the actual device object.

## SYNTAX

```
Expand-Device
	[-InputObject] <Object[]>
	[-Fetch]
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

### EXAMPLE 3
```
Get-DeviceCollection *test* | Expand-Device
```

Get all the device object (with app in their name).
Note the Expand cmdlet won't do much here except for returning the input objects.

### EXAMPLE 4
```
"12345", "mydevice" | Expand-Device -Fetch
```

Expand the devices and always fetch device managed object if an object is not provided via the pipeline

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

### -Fetch
Fetch the full managed object if only the id or name is provided.

```yaml
Type: SwitchParameter
Parameter Sets: (All)
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

### # Without fetch
### [pscustomobject]@{
###     id = "1234"
###     name = "[id=1234]"
### }
### # With fetch
### [pscustomobject]@{
###     id = "1234"
###     name = "mydevice"
### }
## NOTES
If the function calling the Expand-Device has a "Force" parameter and it is set to True, then Expand-Device will not fetch the device managed object
from the server.
Instead it will return an object with only the id and name set (and the name will be set to [id={}]).
This is to save the
number of calls to the server as usually the ID is the item you need to use in subsequent calls.

If the given object is already an device object, then it is added with no additional lookup

The following cases describe when the managed object is fetched from the server and when not.

Cases when the managed object IS fetched from the server
* Calling function does not have -Force on its function, and the user does not use it.
OR
* OR User provides input a string which does not only contain digits
* OR User sets the -Fetch parameter on Expand-Device

Cases when the managed object IS NOT fetched from the server
* User passes an ID like object to Expand-Device
* AND -Force is not used on the calling function
* AND user does not use -Fetch when calling Expand-Device

## RELATED LINKS
