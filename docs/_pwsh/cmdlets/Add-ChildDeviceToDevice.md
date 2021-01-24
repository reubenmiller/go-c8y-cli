---
category: Devices
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Add-ChildDeviceToDevice
---

# Add-ChildDeviceToDevice

## SYNOPSIS
Create a child device reference

## SYNTAX

```
Add-ChildDeviceToDevice
	[-Device] <Object[]>
	[-NewChild] <Object[]>
	[[-ProcessingMode] <String>]
	[[-Template] <String>]
	[[-TemplateVars] <String>]
	[-Raw]
	[[-OutputFile] <String>]
	[-NoProxy]
	[[-Session] <String>]
	[[-TimeoutSec] <Double>]
	[-Force]
	[-WhatIf]
	[-Confirm]
	[<CommonParameters>]
```

## DESCRIPTION
Create a child device reference

## EXAMPLES

### EXAMPLE 1
```
Add-ChildDeviceToDevice -Device $Device.id -NewChild $ChildDevice.id
```

Assign a device as a child device to an existing device

### EXAMPLE 2
```
Get-ManagedObject -Id $ChildDevice.id | Add-ChildDeviceToDevice -Device $Device.id
```

Assign a device as a child device to an existing device (using pipeline)

## PARAMETERS

### -Device
Device.
(required)

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -NewChild
New child device (required)

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: True
Position: 2
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -ProcessingMode
Cumulocity processing mode

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

### -Template
Template (jsonnet) file to use to create the request body.

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

### -TemplateVars
Variables to be used when evaluating the Template.
Accepts a file path, json or json shorthand, i.e.
"name=peter"

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

### System.Object
## NOTES

## RELATED LINKS
