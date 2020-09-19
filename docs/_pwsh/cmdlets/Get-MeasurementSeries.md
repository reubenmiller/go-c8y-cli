---
category: Measurements
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Get-MeasurementSeries
---

# Get-MeasurementSeries

## SYNOPSIS
Get a collection of measurements based on filter parameters

## SYNTAX

```
Get-MeasurementSeries
	[[-Device] <Object[]>]
	[[-Series] <String[]>]
	[[-AggregationType] <String>]
	[[-DateFrom] <String>]
	[[-DateTo] <String>]
	[-Raw]
	[[-OutputFile] <String>]
	[-NoProxy]
	[[-Session] <String>]
	[[-TimeoutSec] <Double>]
	[-WhatIf]
	[-Confirm]
	[<CommonParameters>]
```

## DESCRIPTION
Get a collection of measurements based on filter parameters

## EXAMPLES

### EXAMPLE 1
```
Get-MeasurementSeries -Device $Device.id -Series "c8y_Temperature.T" -DateFrom "1970-01-01" -DateTo "0s"
```

Get a list of measurements for a particular device

### EXAMPLE 2
```
Get-MeasurementSeries -Device $Measurement2.source.id -Series "c8y_Temperature.T" -DateFrom "1970-01-01" -DateTo "0s"
```

Get measurement series c8y_Temperature.T on a device

### EXAMPLE 3
```
Get-DeviceCollection -Name $Device.name | Get-MeasurementSeries -Series "c8y_Temperature.T"
```

Get measurement series from a device (using pipeline)

## PARAMETERS

### -Device
Device ID

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Series
measurement type and series name, e.g.
c8y_AccelerationMeasurement.acceleration

```yaml
Type: String[]
Parameter Sets: (All)
Aliases:

Required: False
Position: 2
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -AggregationType
Fragment name from measurement.

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 3
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -DateFrom
Start date or date and time of measurement occurrence.
Defaults to last 7 days

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 4
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -DateTo
End date or date and time of measurement occurrence.
Defaults to the current time

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 5
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Raw
Show the full (raw) response from Cumulocity including pagination information

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

### -OutputFile
Write the response to file

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 6
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -NoProxy
Ignore any proxy settings when running the cmdlet

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

### -Session
Specifiy alternative Cumulocity session to use when running the cmdlet

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 7
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -TimeoutSec
TimeoutSec timeout in seconds before a request will be aborted

```yaml
Type: Double
Parameter Sets: (All)
Aliases:

Required: False
Position: 8
Default value: 0
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

### System.Object
## NOTES

## RELATED LINKS
