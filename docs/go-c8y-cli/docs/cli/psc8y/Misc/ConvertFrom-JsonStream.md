---
category: Misc
external help file: PSc8y-help.xml
id: ConvertFrom-JsonStream
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Misc/convertfrom-jsonstream
title: ConvertFrom-JsonStream
---



## SYNOPSIS
Convert json text to powershell objects as the objects are piped to it

## SYNTAX

```
ConvertFrom-JsonStream
	[-InputObject] <Object[]>
	[-Depth <Int32>]
	[-AsHashtable]
	[<CommonParameters>]
```

## DESCRIPTION
The cmdlet will convert each input as a separate json line, and it will convert it as soon as it is
received in the pipeline (instead of waiting for the entire input). 

Each input should contain a single json object.

## EXAMPLES

### EXAMPLE 1
```
Get-DeviceCollection | Get-Device -AsPSObject:$false | ConvertFrom-JsonStream
```

Convert the pipeline objects to json as they come through the pipeline.

## PARAMETERS

### -InputObject
Input json lines

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: True (ByValue)
Accept wildcard characters: False
```

### -Depth
Maximum object depth to allow

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: 100
Accept pipeline input: False
Accept wildcard characters: False
```

### -AsHashtable
Convert json to a hashtable instead of a PSCustom Object

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
ConvertFrom-Json can not be used for streamed json as it waits to receive all piped input before it starts trying
to convert the json to PowerShell objects.

## RELATED LINKS
