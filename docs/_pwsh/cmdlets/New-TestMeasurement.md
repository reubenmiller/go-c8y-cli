---
category: Measurements
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: New-TestMeasurement
---

# New-TestMeasurement

## SYNOPSIS
Create a new test measurement

## SYNTAX

```
New-TestMeasurement
	[[-Device] <Object>]
	[[-ValueFragmentType] <String>]
	[[-ValueFragmentSeries] <String>]
	[[-Type] <String>]
	[[-Value] <Double>]
	[[-Unit] <String>]
	[-Force]
	[-WhatIf]
	[-Confirm]
	[<CommonParameters>]
```

## DESCRIPTION
{{ Fill in the Description }}

## EXAMPLES

### Example 1
```powershell
PS C:\> {{ Add example code here }}
```

{{ Add example description here }}

## PARAMETERS

### -Device
{{ Fill Device Description }}

```yaml
Type: Object
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -ValueFragmentType
Value fragment type

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 2
Default value: C8y_Temperature
Accept pipeline input: False
Accept wildcard characters: False
```

### -ValueFragmentSeries
Value fragment series

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 3
Default value: T
Accept pipeline input: False
Accept wildcard characters: False
```

### -Type
Type

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 4
Default value: C8yTemperatureReading
Accept pipeline input: False
Accept wildcard characters: False
```

### -Value
Value

```yaml
Type: Double
Parameter Sets: (All)
Aliases:

Required: False
Position: 5
Default value: 1.2345
Accept pipeline input: False
Accept wildcard characters: False
```

### -Unit
Unit.
i.e.
°C, m/s

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 6
Default value: °C
Accept pipeline input: False
Accept wildcard characters: False
```

### -Force
{{ Fill Force Description }}

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
