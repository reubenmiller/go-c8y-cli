---
category: User Groups
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: New-TestGroup
---

# New-TestGroup

## SYNOPSIS
Create a test user group

## SYNTAX

```
New-TestGroup
	[[-Name] <String>]
	[-Force]
	[<CommonParameters>]
```

## DESCRIPTION
Create a new test user group using a random name

## EXAMPLES

### EXAMPLE 1
```
New-TestGroup -Name mygroup
```

Create a new user group with the prefix "mygroup".
A random postfix will be added to it

## PARAMETERS

### -Name
Name of the user group.
A random postfix will be added to it to make it unique

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: Testgroup
Accept pipeline input: False
Accept wildcard characters: False
```

### -Force
Don't prompt for confirmation

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

## RELATED LINKS
