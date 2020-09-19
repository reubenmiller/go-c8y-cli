---
category: Agents
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: New-TestAgent
---

# New-TestAgent

## SYNOPSIS
Create a new test agent representation in Cumulocity

## SYNTAX

```
New-TestAgent
	[[-Name] <String>]
	[-Force]
	[<CommonParameters>]
```

## DESCRIPTION
Create a new test agent with a randomized name.
Useful when performing mockups or prototyping.

The agent will have both the `c8y_IsDevice` and `com_cumulocity_model_Agent` fragments set.

## EXAMPLES

### EXAMPLE 1
```
New-TestAgent
```

Create a test agent

### EXAMPLE 2
```
1..10 | Foreach-Object { New-TestAgent -Force }
```

Create 10 test agents all with unique names

## PARAMETERS

### -Name
Agent name prefix which is added before the randomized string

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: Testagent
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
