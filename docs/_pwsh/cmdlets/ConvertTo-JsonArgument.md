---
category: Misc.
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: ConvertTo-JsonArgument
---

# ConvertTo-JsonArgument

## SYNOPSIS
Convert a powershell hashtable/object to a json escaped string

## SYNTAX

```
ConvertTo-JsonArgument
	[-Data] <Object>
	[<CommonParameters>]
```

## DESCRIPTION
Helper function is used when passing Powershell hashtable or PSCustomObjects to
the c8y binary.
Before the c8y cli binary can accept it, it must be converted to json.

The necessary character escaping of literal backslashed `\\` will be done automatically.

## EXAMPLES

### EXAMPLE 1
```
ConvertTo-JsonArgument @{ myValue = "1" }
```

Converts the hashtable to an escaped json string

```json
{\"myValue\":\"1\"}
```

## PARAMETERS

### -Data
Input object to be converted

```yaml
Type: Object
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

## NOTES

## RELATED LINKS
