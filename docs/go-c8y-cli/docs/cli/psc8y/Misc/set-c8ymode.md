---
category: Misc
external help file: PSc8y-help.xml
id: Set-c8yMode
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Misc/set-c8ymode
title: Set-c8yMode
---



## SYNOPSIS
Set cli mode temporarily

## SYNTAX

```
Set-c8yMode
	[-Mode] <String>
	[<CommonParameters>]
```

## DESCRIPTION
Set cli mode temporarily

## EXAMPLES

### EXAMPLE 1
```
Set-c8yMode -Mode dev
```

Enable development mode (all command enabled) temporarily.
The active session file will not be updated

## PARAMETERS

### -Mode
Mode

```yaml
Type: String
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
