---
category: Devices
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: New-TestDevice
---

# New-TestDevice

## SYNOPSIS
Create a new test device representation in Cumulocity

## SYNTAX

```
New-TestDevice
	[[-Name] <String>]
	[-AsAgent]
	[-Force]
	[-WhatIf]
	[-Confirm]
	[<CommonParameters>]
```

## DESCRIPTION
Create a new test device with a randomized name.
Useful when performing mockups or prototyping.

The agent will have both the `c8y_IsDevice` fragments set.

## EXAMPLES

### EXAMPLE 1
```
New-TestDevice
```

Create a test device

### EXAMPLE 2
```
1..10 | Foreach-Object { New-TestDevice -Force }
```

Create 10 test devices all with unique names

### EXAMPLE 3
```
1..10 | Foreach-Object { New-TestDevice -AsAgent -Force }
```

Create 10 test devices (with agent functionality) all with unique names

## PARAMETERS

### -Name
Device name prefix which is added before the randomized string

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: Testdevice
Accept pipeline input: False
Accept wildcard characters: False
```

### -AsAgent
Add agent fragment to the device

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

### -WhatIf
Shows what would happen if the cmdlet runs.
The cmdlet is not run.

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases: wi

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Confirm
Prompts you for confirmation before running the cmdlet.

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases: cf

Required: False
Position: Named
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
