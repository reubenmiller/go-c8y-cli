---
category: Misc
external help file: PSc8y-help.xml
id: Open-Website
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Misc/open-website
title: Open-Website
---



## SYNOPSIS
Open a browser to the cumulocity website

## SYNTAX

### Device (Default)
```
Open-Website
	[[-Device] <Object[]>]
	[[-Page] <String>]
	[-Browser <String>]
	[<CommonParameters>]
```

### Application
```
Open-Website
	[[-Application] <String>]
	[-Browser <String>]
	[<CommonParameters>]
```

## DESCRIPTION
Opens the default web browser to the Cumulocity application or directly to a device page in the Device Management application

## EXAMPLES

### EXAMPLE 1
```
Open-Website -Application "cockpit"
```

Open the cockpit application

### EXAMPLE 2
```
Open-Website -Device myDevice01
```

Open the devicemanagement to the device (default) control page for myDevice01

### EXAMPLE 3
```
Open-Website -Device myDevice01 -Page alarms
```

Open the devicemanagement to the device alarm page for myDevice01

## PARAMETERS

### -Application
Application to open

```yaml
Type: String
Parameter Sets: Application
Aliases:

Required: False
Position: 1
Default value: Cockpit
Accept pipeline input: False
Accept wildcard characters: False
```

### -Device
Name of the device to open in devicemanagement.
Only the first matching device will be used to open the c8y website.

```yaml
Type: Object[]
Parameter Sets: Device
Aliases:

Required: False
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Page
Device page to open

```yaml
Type: String
Parameter Sets: Device
Aliases:

Required: False
Position: 2
Default value: Control
Accept pipeline input: False
Accept wildcard characters: False
```

### -Browser
Browser to use to open the webpage

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: Chrome
Accept pipeline input: False
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

## NOTES
When running on Linux, it relies on xdg-open.
If it is not found, then only the URL will be printed to the console.
The user can then try to open the URL by clicking on the link if they are using a modern terminal which supports url links.

## RELATED LINKS
