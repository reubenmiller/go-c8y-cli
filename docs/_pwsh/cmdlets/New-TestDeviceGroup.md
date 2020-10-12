---
category: Devices
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: New-TestDeviceGroup
---

# New-TestDeviceGroup

## SYNOPSIS
Create a new test device group

## SYNTAX

```
New-TestDeviceGroup
	[[-Name] <String>]
	[-Type <String>]
	[-Template <String>]
	[-TemplateVars <String>]
	[-Force]
	[-WhatIf]
	[-Confirm]
	[<CommonParameters>]
```

## DESCRIPTION
Create a new test device group with a randomized name.
Useful when performing mockups or prototyping.

## EXAMPLES

### EXAMPLE 1
```
New-TestDeviceGroup
```

Create a test device group

### EXAMPLE 2
```
1..10 | Foreach-Object { New-TestDeviceGroup -Force }
```

Create 10 test device groups all with unique names

## PARAMETERS

### -Name
Device group name prefix which is added before the randomized string

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: Testgroup
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Type
Group type.
Only device groups of type `Group` are visible as root folders in the UI

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: Group
Accept pipeline input: False
Accept wildcard characters: False
```

### -Template
Template (jsonnet) file to use to create the request body.

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -TemplateVars
Variables to be used when evaluating the Template.
Accepts json or json shorthand, i.e.
"name=peter"

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
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
