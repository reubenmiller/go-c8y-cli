---
category: Sessions
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Get-Session
---

# Get-Session

## SYNOPSIS
Get the active Cumulocity Session

## SYNTAX

```
Get-Session
	[[-Session] <String>]
	[<CommonParameters>]
```

## DESCRIPTION
Get the details about the active Cumulocity session which is used by all cmdlets

## EXAMPLES

### EXAMPLE 1
```
Get-Session
```

Get the current Cumulocity session

## PARAMETERS

### -Session
Specifiy alternative Cumulocity session to use when running the cmdlet

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

### None
## NOTES

## RELATED LINKS
