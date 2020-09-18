---
category: Operations
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Wait-Operation
---

# Wait-Operation

## SYNOPSIS
Wait for an operation to be completed (i.e.
either in the SUCCESS or FAILED status)

## SYNTAX

```
Wait-Operation
	[-Id] <String>
	[-TimeoutSec <Int32>]
	[<CommonParameters>]
```

## DESCRIPTION
{{ Fill in the Description }}

## EXAMPLES

### EXAMPLE 1
```
Wait-Operation 1234567
```

Wait for the operation id

### EXAMPLE 2
```
Wait-Operation 1234567 -TimeoutSec 30
```

Wait for the operation id, and timeout after 30 seconds

## PARAMETERS

### -Id
Operation id or object to wait for

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -TimeoutSec
Timeout in seconds.
Defaults to 30 seconds.
i.e.
how long should it wait for the operation to be processed

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: 30
Accept pipeline input: False
Accept wildcard characters: False
```

### CommonParameters
This cmdlet supports the common parameters: -Debug, -ErrorAction, -ErrorVariable, -InformationAction, -InformationVariable, -OutVariable, -OutBuffer, -PipelineVariable, -Verbose, -WarningAction, and -WarningVariable. For more information, see [about_CommonParameters](http://go.microsoft.com/fwlink/?LinkID=113216).

## INPUTS

## OUTPUTS

## NOTES

## RELATED LINKS
