---
category: Misc.
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Format-Date
---

# Format-Date

## SYNOPSIS
Gets a Cumulocity (ISO-8601) formatted DateTime string in the specified timezone

## SYNTAX

```
Format-Date
	[[-InputObject] <DateTime[]>]
	[-TimeZone <TimeZoneInfo>]
	[<CommonParameters>]
```

## DESCRIPTION
All Cumulocity REST API calls that require a date, must be in the ISO-8601 format.
This function
allows the user to easily generate the correct format including the correct timezone information.

## EXAMPLES

### EXAMPLE 1
```
Format-Date
```

Get current datetime (now) as an ISO8601 formatted string

### EXAMPLE 2
```
[TimeZoneInfo]::GetSystemTimeZones() | Foreach-Object { Format-Date -Timezone $_ }
```

Get current datetime (now) as an ISO8601 formatted string in each of the timezones

## PARAMETERS

### -InputObject
DateTime to be converted to ISO-8601 format.
Accepts piped input

```yaml
Type: DateTime[]
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: @()
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -TimeZone
Timezone to use when converting the DateTime object.
Defaults to Local System Timezone

```yaml
Type: TimeZoneInfo
Parameter Sets: (All)
Aliases:

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

### String
## NOTES
The standard powershell Get-Date does not have any timezone information.

## RELATED LINKS
