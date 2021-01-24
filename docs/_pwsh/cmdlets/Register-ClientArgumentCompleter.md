---
category: Client Helpers
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Register-ClientArgumentCompleter
---

# Register-ClientArgumentCompleter

## SYNOPSIS
Register PSc8y argument completions for a specific cmdlet

## SYNTAX

```
Register-ClientArgumentCompleter
	[-Name] <String[]>
	[-Force]
	[<CommonParameters>]
```

## DESCRIPTION
The cmdlet enables support for argument completion which are used within PSc8y in other modules.

## EXAMPLES

### EXAMPLE 1
```
Register-ClientArgumentCompleter -Name "Get-MyCustomCommand"
```

Register PSc8y argument completion for supported parameters for a custom function called "Get-MyCustomCommand"

### EXAMPLE 2
```
Register-ClientArgumentCompleter -Name "New-CustomManagedObject" -Force
```

Force the registration of argument completers on a function which uses dynamic parameters

## PARAMETERS

### -Name
Command Name

```yaml
Type: String[]
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Force
Force the registration of all parameter (required when a cmdlet has Dynamic Parameters)

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

## NOTES
The following argument completions are supports

* `-Session` - Session selection completion
* `-Template` - Template file completion

## RELATED LINKS
