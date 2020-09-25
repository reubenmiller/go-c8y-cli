---
category: Sessions
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Set-Session
---

# Set-Session

## SYNOPSIS
Set/activate a Cumulocity Session.

## SYNTAX

### ByInteraction (Default)
```
Set-Session
	[[-Filter] <String[]>]
	[-UseEnvironment]
	[<CommonParameters>]
```

### ByFile
```
Set-Session
	[[-File] <String>]
	[-UseEnvironment]
	[<CommonParameters>]
```

## DESCRIPTION
By default the user will be prompted to select from Cumulocity sessions found in their home folder under .cumulocity

Filtering the list is always 

"customer dev" will be split in to two search terms, "customer" and "dev", and only results which contain these two search
terms will be includes in the results.
The search is applied to the following fields of the session:

* index
* filename (basename only)
* host
* tenant
* username

## EXAMPLES

### EXAMPLE 1
```
Set-Session
```

Prompt for a list of Cumulocity sessions to select from

### EXAMPLE 2
```
Set-Session customer
```

Set a session interactively but only include sessions where the details contain "customer" in any of the fields

### EXAMPLE 3
```
Set-Session customer, dev
```

Set a session interactively but only includes session where the details includes "customer" and "dev" in any of the fields

## PARAMETERS

### -File
File containing the Cumulocity session data

```yaml
Type: String
Parameter Sets: ByFile
Aliases: FullName

Required: False
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Filter
Filter list of sessions.
Multiple search terms can be provided.
A string "Contains" operation
is done to match any of the session fields (except password)

```yaml
Type: String[]
Parameter Sets: ByInteraction
Aliases:

Required: False
Position: 1
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -UseEnvironment
Allow loading Cumulocity session setting from environment variables

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

### String
## NOTES

## RELATED LINKS
