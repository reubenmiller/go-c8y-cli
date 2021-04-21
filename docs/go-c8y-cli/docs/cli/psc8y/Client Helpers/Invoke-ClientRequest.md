---
category: Client Helpers
external help file: PSc8y-help.xml
id: Invoke-ClientRequest
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Client Helpers/invoke-clientrequest
title: Invoke-ClientRequest
---



## SYNOPSIS
Send a rest request using the c8y

## SYNTAX

```
Invoke-ClientRequest
	[-Uri] <String>
	[-Method <WebRequestMethod>]
	[-Headers <Hashtable>]
	[-InFile <String>]
	[-QueryParameters <Hashtable>]
	[-ContentType <String>]
	[-Accept <String>]
	[-HostName <String>]
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
	[-Timeout <Double>]
	[-Session <String>]
	[-Output <String>]
	[-View <String>]
	[-AsHashTable]
	[-AsPSObject]
	[-Flatten]
	[-Compress]
	[-Pretty]
	[-NoColor]
	[-Color]
	[-Confirm]
	[-ConfirmText <String>]
	[-WithError]
	[-SilentStatusCodes <String>]
	[-Dry]
	[-DryFormat <String>]
	[-Workers <Int32>]
	[-Delay <Int32>]
	[-MaxJobs <Int32>]
	[-Progress]
	[-NoLog]
	[-LogMessage <String>]
	[-Select <String[]>]
	[<CommonParameters>]
```

## DESCRIPTION
Send a custom rest request to Cumulocity using all of the options found on other command lets.
This is useful if you are extending PSc8y and want to send custom microservice requests, or
send requests which are not yet provided in the PSc8y module.

Example:

The following function sends a POST request to predefined microservice endpoint.
It accepts an input Body argument which will be used in the request.

The response is also converted from raw json (string) to Powershell objects so that advanced
filtering can be done on the response (i.e.
using `Where-Object`)

```powershell
Function Invoke-MyMicroserviceEndpoint {
    [cmdletbinding()]
    Param(
        [hashtable] $Body
    )

    $options = @{
        Method = "POST"
        Uri = "/service/mymicroservice"
        Data = $Body
    }

    # Send request
    $response = Invoke-ClientRequest @options
    
    # Convert response from json to Powershell objects
    ConvertFrom-Json $response
}
```

## EXAMPLES

### EXAMPLE 1
```
Invoke-ClientRequest -Uri "/inventory/managedObjects" -Method "post" -Data "name=test"
```

Create a new managed object with the name "test"

### EXAMPLE 2
```
Invoke-ClientRequest -Uri "/alarm/alarms" -QueryParameters @{ pageSize = "100" }
```

Get a list of alarms with page size of 100

### EXAMPLE 3
```
Invoke-ClientRequest -Uri "/alarm/alarms?pageSize=100"
```

Get a list of alarms with page size of 100

### EXAMPLE 4
```
Invoke-ClientRequest -Uri "/inventory/managedObjects" -Method "post" -Data "name=test" -Headers @{ Custom-Value = "myValue"}
```

Create a new managed object but add a custom accept header value

## PARAMETERS

### -Uri
Uri (or partial uri).
i.e.
/application/applications

```yaml
Type: String
Parameter Sets: (All)
Aliases:

Required: True
Position: 1
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Method
Rest Method.
Defaults to GET

```yaml
Type: WebRequestMethod
Parameter Sets: (All)
Aliases:
Accepted values: Default, Get, Head, Post, Put, Delete, Trace, Options, Merge, Patch

Required: False
Position: Named
Default value: GET
Accept pipeline input: False
Accept wildcard characters: False
```

### -Headers
Add custom headers to the rest request

```yaml
Type: Hashtable
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -InFile
Input file to be uploaded as FormData

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

### -QueryParameters
Uri query parameters

```yaml
Type: Hashtable
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -ContentType
(Body) Content Type

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

### -Accept
Accept header

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

### -HostName
HostName to use which overrides the given host

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

### -AsHashTable
{{ Fill AsHashTable Description }}

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
{{ Fill AsPSObject Description }}

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

### -Color
{{ Fill Color Description }}

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

### -Compress
{{ Fill Compress Description }}

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

### -Confirm
Prompts you for confirmation before running the cmdlet.

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
{{ Fill ConfirmText Description }}

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

### -Data
{{ Fill Data Description }}

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
{{ Fill Delay Description }}

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

### -Dry
{{ Fill Dry Description }}

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
{{ Fill DryFormat Description }}

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

### -Flatten
{{ Fill Flatten Description }}

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
{{ Fill Force Description }}

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
{{ Fill LogMessage Description }}

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
{{ Fill MaxJobs Description }}

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
{{ Fill NoAccept Description }}

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
{{ Fill NoColor Description }}

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
{{ Fill NoLog Description }}

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
{{ Fill NoProxy Description }}

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
{{ Fill Output Description }}

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
{{ Fill OutputFile Description }}

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
{{ Fill OutputFileRaw Description }}

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

### -Pretty
{{ Fill Pretty Description }}

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

### -ProcessingMode
{{ Fill ProcessingMode Description }}

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
{{ Fill Progress Description }}

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
{{ Fill Proxy Description }}

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
{{ Fill Raw Description }}

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
{{ Fill Select Description }}

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
{{ Fill Session Description }}

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

### -SilentStatusCodes
{{ Fill SilentStatusCodes Description }}

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
{{ Fill Template Description }}

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
{{ Fill TemplateVars Description }}

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
{{ Fill Timeout Description }}

```yaml
Type: Double
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -View
{{ Fill View Description }}

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
{{ Fill WithError Description }}

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
{{ Fill Workers Description }}

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

## NOTES

## RELATED LINKS

[https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/rest](https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/rest)

