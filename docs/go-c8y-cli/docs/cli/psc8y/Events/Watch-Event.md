---
category: Events
external help file: PSc8y-help.xml
id: Watch-Event
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Events/watch-event
title: Watch-Event
---



## SYNOPSIS
Watch realtime events

## SYNTAX

```
Watch-Event
	[[-Device] <Object>]
	[[-Duration] <Int32>]
	[[-Count] <Int32>]
	[[-ActionTypes] <String[]>]
	[<CommonParameters>]
```

## DESCRIPTION
Watch realtime events

## EXAMPLES

### EXAMPLE 1
```
Watch-Event -Device 12345
Watch all events for a device
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
Start date or date and time of event occurrence.
(required)

```yaml
Type: Int32
Parameter Sets: (All)
Aliases: DurationSec

Required: False
Position: 2
Default value: 0
Accept pipeline input: False
Accept wildcard characters: False
```

### -Count
End date or date and time of event occurrence.

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

[https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/events_subscribe](https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/events_subscribe)

