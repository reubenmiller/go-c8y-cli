---
category: Alarms
external help file: PSc8y-help.xml
id: Update-AlarmCollection
Module Name: PSc8y
online version: https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/alarms_updateCollection
schema: 2.0.0
slug: /docs/cli/psc8y/Alarms/update-alarmcollection
title: Update-AlarmCollection
---



## SYNOPSIS
Update alarm collection

## SYNTAX

```
Update-AlarmCollection
	[[-Device] <Object[]>]
	[[-Status] <String>]
	[[-Severity] <String>]
	[-Resolved]
	[[-DateFrom] <String>]
	[[-DateTo] <String>]
	[[-NewStatus] <String>]
	[-Data <Object>]
	[-NoAccept]
	[-ProcessingMode <String>]
	[-Force]
	[-Template <String>]
	[-TemplateVars <String>]
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
Update the status of a collection of alarms by using a filter.
Currently only the status of alarms can be changed

## EXAMPLES

### EXAMPLE 1
```
Update-AlarmCollection -Device $Device.id -Status ACTIVE -NewStatus ACKNOWLEDGED
```

Update the status of all active alarms on a device to ACKNOWLEDGED

### EXAMPLE 2
```
Get-Device -Id $Device.id | PSc8y\Update-AlarmCollection -Status ACTIVE -NewStatus ACKNOWLEDGED
```

Update the status of all active alarms on a device to ACKNOWLEDGED (using pipeline)

## PARAMETERS

### -Device
The ManagedObject that the alarm originated from

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

### -Status
The status of the alarm: ACTIVE, ACKNOWLEDGED or CLEARED.
If status was not appeared, new alarm will have status ACTIVE.
Must be upper-case.

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

### -Severity
The severity of the alarm: CRITICAL, MAJOR, MINOR or WARNING.
Must be upper-case.

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

### -Resolved
When set to true only resolved alarms will be removed (the one with status CLEARED), false means alarms with status ACTIVE or ACKNOWLEDGED.

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

### -DateFrom
Start date or date and time of alarm occurrence.

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
End date or date and time of alarm occurrence.

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

### -NewStatus
New status to be applied to all of the matching alarms

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

### -Data
static data to be applied to body.
accepts json or shorthande json, i.e.
--data 'value1=1,my.nested.value=100'

```yaml
Type: Object
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

### -Force
Do not prompt for confirmation.
Ignored when using --confirm

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

### -NoAccept
Ignore Accept header will remove the Accept header from requests, however PUT and POST requests will only see the effect

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

### -ProcessingMode
Cumulocity processing mode

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

### -Template
Body template

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
Body template variables

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

[https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/alarms_updateCollection](https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/alarms_updateCollection)

