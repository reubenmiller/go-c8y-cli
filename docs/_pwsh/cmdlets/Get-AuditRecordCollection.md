---
category: AuditRecords
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Get-AuditRecordCollection
---

# Get-AuditRecordCollection

## SYNOPSIS
Get collection of (user) audits

## SYNTAX

```
Get-AuditRecordCollection
	[[-Source] <Object>]
	[[-Type] <String>]
	[[-User] <String>]
	[[-Application] <String>]
	[[-DateFrom] <String>]
	[[-DateTo] <String>]
	[-Revert]
	[[-PageSize] <Int32>]
	[-WithTotalPages]
	[[-CurrentPage] <Int32>]
	[[-TotalPages] <Int32>]
	[-IncludeAll]
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
Audit records contain information about modifications to other Cumulocity entities.
For example the audit records contain each operation state transition, so they can be used to check when an operation transitioned from PENDING -\> EXECUTING -\> SUCCESSFUL.

## EXAMPLES

### EXAMPLE 1
```
Get-AuditRecordCollection -PageSize 100
```

Get a list of audit records

### EXAMPLE 2
```
Get-AuditRecordCollection -Source $Device2.id
```

Get a list of audit records related to a managed object

### EXAMPLE 3
```
Get-Operation -Id $Operation.id | Get-AuditRecordCollection
```

Get a list of audit records related to an operation

## PARAMETERS

### -Source
Source Id or object containing an .id property of the element that should be detected.
i.e.
AlarmID, or Operation ID.
Note: Only one source can be provided

```yaml
Type: Object
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -Type
Type

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 2
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -User
Username

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

### -Application
Application

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

### -DateFrom
Start date or date and time of audit record occurrence.

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

### -DateTo
End date or date and time of audit record occurrence.

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

### -Revert
Return the newest instead of the oldest audit records.
Must be used with dateFrom and dateTo parameters

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

### -PageSize
Maximum number of results

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: 7
Default value: 0
Accept pipeline input: False
Accept wildcard characters: False
```

### -WithTotalPages
Include total pages statistic

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

### -CurrentPage
Get a specific page result

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: 8
Default value: 0
Accept pipeline input: False
Accept wildcard characters: False
```

### -TotalPages
Maximum number of pages to retrieve when using -IncludeAll

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: 9
Default value: 0
Accept pipeline input: False
Accept wildcard characters: False
```

### -IncludeAll
Include all results

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
Position: 10
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
Position: 11
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
Position: 12
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
