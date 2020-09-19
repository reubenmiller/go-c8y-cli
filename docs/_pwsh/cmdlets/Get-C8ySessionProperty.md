---
category: Sessions
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Get-C8ySessionProperty
---

# Get-C8ySessionProperty

## SYNOPSIS
Get a property from the current c8y session

## SYNTAX

```
Get-C8ySessionProperty
	[-Name] <String>
	[<CommonParameters>]
```

## DESCRIPTION
An interface to read properties from the current c8y session, i.e.
tenant or host.
This is mostly used
internally my other cmdlets in the module to abstract the accessing of such variables in case the environment
variables change in the future (i.e.
$env:C8Y_TENANT or $env:C8Y_HOST).

## EXAMPLES

### EXAMPLE 1
```
Get-C8ySessionProperty tenant
```

Get the tenant name of the current c8y cli session

## PARAMETERS

### -Name
Property name

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

### string
## NOTES

## RELATED LINKS
