---
category: Operations
external help file: PSc8y-help.xml
id: Get-BulkOperationOperationCollection
Module Name: PSc8y
online version: https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/bulkoperations_listOperations
schema: 2.0.0
slug: /docs/cli/psc8y/Operations/get-bulkoperationoperationcollection
title: Get-BulkOperationOperationCollection
---



## SYNOPSIS
Get operations collection

## SYNTAX

```
Get-BulkOperationOperationCollection
	[-Id] <Object[]>
	[[-DateFrom] <String>]
	[[-DateTo] <String>]
	[[-Status] <String>]
	[-Revert]
	[-PageSize <Int32>]
	[-WithTotalPages]
	[-CurrentPage <Int32>]
	[-TotalPages <Int32>]
	[-IncludeAll]
	[-Raw]
	[-OutputFile <String>]
	[-OutputFileRaw <String>]
	[-Proxy]
	[-NoProxy]
	[-Timeout <String>]
	[-Session <String>]
	[-SessionUsername <String>]
	[-SessionPassword <String>]
	[-Output <String>]
	[-View <String>]
	[-AsHashTable]
	[-AsPSObject]
	[-Flatten]
	[-Compact]
	[-NoColor]
	[-Cache]
	[-NoCache]
	[-CacheTTL <String>]
	[-Insecure]
	[-Help]
	[-Examples]
	[-Confirm]
	[-ConfirmText <String>]
	[-WithError]
	[-SilentStatusCodes <String>]
	[-SilentExit]
	[-Dry]
	[-DryFormat <String>]
	[-Workers <Int32>]
	[-Delay <String>]
	[-DelayBefore <String>]
	[-MaxJobs <Int32>]
	[-Progress]
	[-AbortOnErrors <Int32>]
	[-NoLog]
	[-LogMessage <String>]
	[-Select <String[]>]
	[-Filter <String[]>]
	[-Header <String[]>]
	[-CustomQueryParam <String[]>]
	[<CommonParameters>]
```

## DESCRIPTION
Get a collection of operations related to a bulk operation

## EXAMPLES

### EXAMPLE 1
```
Get-BulkOperationOperationCollection -Id 10 -Status PENDING
```

Get a list of pending operations from bulk operation with id 10

### EXAMPLE 2
```
Get-BulkOperationCollection | Where-Object { $_.status -eq "IN_PROGRESS" } | Get-BulkOperationOperationCollection -Status PENDING
```

Get all pending operations from all bulk operations which are still in progress (using pipeline)

### EXAMPLE 3
```
Get-BulkOperationCollection | Get-BulkOperationOperationCollection -status EXECUTING --dateTo "-10d" | Update-Operation -Status FAILED -FailureReason "Manually cancelled stale operation"
```

Check all bulk operations if they have any related operations still in executing state and were created more than 10 days ago, then cancel it with a custom message

## PARAMETERS

### -Id
Bulk operation id.
(required)

```yaml
Type: Object[]
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: True (ByPropertyName, ByValue)
Accept wildcard characters: False
```

### -DateFrom
Start date or date and time of operation.

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

### -DateTo
End date or date and time of operation.

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

### -Status
Operation status, can be one of SUCCESSFUL, FAILED, EXECUTING or PENDING.

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

### -Revert
Sort operations newest to oldest.
Must be used with dateFrom and/or dateTo parameters

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

### -AbortOnErrors
Abort batch when reaching specified number of errors

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -AsHashTable
Return output as PowerShell Hashtables

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -AsPSObject
Return output as PowerShell PSCustomObjects

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Cache
Enable cached responses

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -CacheTTL
Cache time-to-live (TTL) as a duration, i.e.
60s, 2m

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

### -Compact
Compact instead of pretty-printed output when using json output.
Pretty print is the default if output is the terminal

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases: Compress

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Confirm
Prompt for confirmation

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -ConfirmText
Custom confirmation text

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

### -CurrentPage
Current page which should be returned

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -CustomQueryParam
add custom URL query parameters.
i.e.
--customQueryParam 'withCustomOption=true,myOtherOption=myvalue'

```yaml
Type: String[]
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Delay
delay after each request.
It accepts a duration, i.e.
1ms, 0.5s, 1m etc.

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

### -DelayBefore
delay before each request.
It accepts a duration, i.e.
1ms, 0.5s, 1m etc.

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

### -Dry
Dry run.
Don't send any data to the server

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -DryFormat
Dry run output format.
i.e.
json, dump, markdown or curl

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

### -Examples
Show examples for the current command

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Filter
Apply a client side filter to response before returning it to the user

```yaml
Type: String[]
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Flatten
flatten json output by replacing nested json properties with properties where their names are represented by dot notation

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Header
custom headers.
i.e.
--header 'Accept: value, AnotherHeader: myvalue'

```yaml
Type: String[]
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Help
Show command help

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -IncludeAll
Include all results by iterating through each page

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Insecure
Allow insecure server connections when using SSL

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -LogMessage
Add custom message to the activity log

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

### -MaxJobs
Maximum number of jobs.
0 = unlimited (use with caution!)

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -NoCache
Force disabling of cached responses (overwrites cache setting)

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -NoColor
Don't use colors when displaying log entries on the console

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -NoLog
Disables the activity log for the current command

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -NoProxy
Ignore the proxy settings

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Output
Output format i.e.
table, json, csv, csvheader

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

### -OutputFile
Save JSON output to file (after select/view)

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

### -OutputFileRaw
Save raw response to file (before select/view)

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

### -PageSize
Maximum results per page

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Progress
Show progress bar.
This will also disable any other verbose output

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Proxy
Proxy setting, i.e.
http://10.0.0.1:8080

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Raw
Show raw response.
This mode will force output=json and view=off

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Select
Comma separated list of properties to return.
wildcards and globstar accepted, i.e.
--select 'id,name,type,**.serialNumber'

```yaml
Type: String[]
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Session
Session configuration

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

### -SessionPassword
Override session password

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

### -SessionUsername
Override session username.
i.e.
peter or t1234/peter (with tenant)

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

### -SilentExit
Silent status codes do not affect the exit code

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -SilentStatusCodes
Status codes which will not print out an error message

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

### -Timeout
Request timeout.
It accepts a duration, i.e.
1ms, 0.5s, 1m etc.

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

### -TotalPages
Total number of pages to get

```yaml
Type: Int32
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -View
Use views when displaying data on the terminal.
Disable using --view off

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

### -WithError
Errors will be printed on stdout instead of stderr

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -WithTotalPages
Request Cumulocity to include the total pages in the response statitics under .statistics.totalPages

```yaml
Type: SwitchParameter
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Workers
Number of workers

```yaml
Type: Int32
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

### System.Object
## NOTES

## RELATED LINKS

[https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/bulkoperations_listOperations](https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/bulkoperations_listOperations)

