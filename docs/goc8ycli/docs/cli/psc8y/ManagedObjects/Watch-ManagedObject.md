---
category: ManagedObjects
external help file: PSc8y-help.xml
id: Watch-ManagedObject
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/ManagedObjects/watch-managedobject
title: Watch-ManagedObject
---



## SYNOPSIS
Watch realtime managedObjects

## SYNTAX

```
Watch-ManagedObject
	[[-Device] <Object>]
	[[-Duration] <String>]
	[[-Count] <Int32>]
	[[-ActionTypes] <String[]>]
	[<CommonParameters>]
```

## DESCRIPTION
Watch realtime managedObjects

## EXAMPLES

### EXAMPLE 1
```
Watch-ManagedObject -Device 12345
Watch all managedObjects for a device
```

## PARAMETERS

### -Device
Device ID

```yaml
Type: Object
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Duration
Duration to subscribe for.
It accepts a duration, i.e.
1ms, 0.5s, 1m etc.

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 2
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Count
End date or date and time of managedObject occurrence.

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

### -ActionTypes
Filter by realtime action types, i.e.
CREATE,UPDATE,DELETE

```yaml
Type: String[]
Parameter Sets: (All)
Aliases:

Required: False
Position: 4
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

### System.Object
## NOTES

## RELATED LINKS

[https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_subscribe](https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_subscribe)

