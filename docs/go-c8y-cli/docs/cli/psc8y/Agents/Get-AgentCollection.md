---
category: Agents
external help file: PSc8y-help.xml
id: Get-AgentCollection
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Agents/get-agentcollection
title: Get-AgentCollection
---



## SYNOPSIS
Get a collection of agents

## SYNTAX

```
Get-AgentCollection
	[[-Name] <String>]
	[[-Type] <String>]
	[[-FragmentType] <String>]
	[[-Owner] <String>]
	[[-Availability] <String>]
	[[-LastMessageDateFrom] <String>]
	[[-LastMessageDateTo] <String>]
	[[-Group] <String>]
	[[-Query] <String>]
	[[-QueryTemplate] <String>]
	[[-OrderBy] <String>]
	[-WithParents]
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
Get a collection of agent in the current tenant

## EXAMPLES

### EXAMPLE 1
```
Get-AgentCollection -Name *sensor*
```

Get all agents with "sensor" in their name

### EXAMPLE 2
```
Get-AgentCollection -Name *sensor* -Type *c8y_* -PageSize 100
```

Get the first 100 agents with "sensor" in their name and has a type matching "c8y_"

### EXAMPLE 3
```
Get-AgentCollection -Query "lastUpdated.date gt '2020-01-01T00:00:00Z'"
```

Get a list of agents which have been updated more recently than 2020-01-01

## PARAMETERS

### -Name
Agent name.
Wildcards accepted

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 1
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Type
Agent type.

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

### -FragmentType
Agent fragment type.

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

### -Owner
Agent owner.

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

### -Availability
Availability.

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

### -LastMessageDateFrom
LastMessageDateFrom - c8y_Availability.lastMessage filter

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

### -LastMessageDateTo
LastMessageDateTo - c8y_Availability.lastMessage filter

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

### -Group
Group.

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 8
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Query
Query.

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: False
Position: 9
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -QueryTemplate
QueryTemplate.

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

### -OrderBy
Order results by a specific field.
i.e.
"name", "_id desc" or "creationTime.date asc".

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

### -WithParents
include a flat list of all parents and grandparents of the given object

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
Request Cumulocity to include the total pages in the response statistics under .statistics.totalPages

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

[https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/agents_list](https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/agents_list)

