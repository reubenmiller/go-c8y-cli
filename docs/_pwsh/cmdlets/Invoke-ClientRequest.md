---
category: Client Helpers
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: Invoke-ClientRequest
---

# Invoke-ClientRequest

## SYNOPSIS
Send a rest request using the c8y

## SYNTAX

```
Invoke-ClientRequest
	[-Uri] <String>
	[-HostName <String>]
	[-Method <WebRequestMethod>]
	[-Data <Object>]
	[-Headers <Hashtable>]
	[-InFile <String>]
	[-QueryParameters <Hashtable>]
	[-ContentType <String>]
	[-Accept <String>]
	[-IgnoreAcceptHeader]
	[-TimeoutSec <Double>]
	[-Pretty]
	[-Raw]
	[-OutputFile <String>]
	[-NoProxy]
	[-Session <String>]
	[-UseEnvironment]
	[-WhatIf]
	[-Confirm]
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
    [cmdletbinding(
        SupportsShouldProcess = $true
    )]
    Param(
        [hashtable] $Body
    )

    $options = @{
        Method = "POST"
        Uri = "/service/mymicroservice"
        Data = $Body

        # Add these to support -WhatIf and -Verbose parameters
        WhatIfPreference = $WhatIfPreference `
        VerbosePreference = $VerbosePreference
    }

    # Send request
    $response = Invoke-ClientRequest @options
    
    # Convert response from json to Powershell objects
    ConvertFrom-Json $response -Depth 100
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
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Data
Request body

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

### -IgnoreAcceptHeader
Ignore the accept header

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

### -TimeoutSec
Timeout in seconds

```yaml
Type: Double
Parameter Sets: (All)
Aliases:

Required: False
Position: Named
Default value: 0
Accept pipeline input: False
Accept wildcard characters: False
```

### -Pretty
Pretty print json response

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
Include raw response including pagination information

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
Outputfile

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

### -NoProxy
NoProxy

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
Session path

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

### -UseEnvironment
Allow loading Cumulocity session setting from environment variables

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

## NOTES

## RELATED LINKS
