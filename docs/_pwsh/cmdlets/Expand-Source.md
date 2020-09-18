---
category: Misc.
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Expand-Source
---

# Expand-Source

## SYNOPSIS
Expand a list of source ids.

## SYNTAX

```
Expand-Source
	[-InputObject] <Object[]>
	[<CommonParameters>]
```

## DESCRIPTION
1.
Look for a source.id property
2.
Look for a deviceId property
3.
Look for a id property
4.
Check if the given is a string or int and is integer like

## EXAMPLES

### EXAMPLE 1
```
Expand-Source 12345
```

Normalize a list of ids

### EXAMPLE 2
```
"12345", "56789" | Expand-Source
```

Normalize a list of ids

## PARAMETERS

### -InputObject
List of ids

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

## RELATED LINKS
