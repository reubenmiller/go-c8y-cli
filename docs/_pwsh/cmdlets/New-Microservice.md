---
category: Microservices
external help file: PSc8y-help.xml
layout: powershell
Module Name: PSc8y
online version:
schema: 2.0.0
title: New-Microservice
---

# New-Microservice

## SYNOPSIS
New microservice

## SYNTAX

```
New-Microservice
	[-File] <String>
	[[-Name] <String>]
	[[-Key] <String>]
	[[-Availability] <String>]
	[[-ContextPath] <String>]
	[[-ResourcesUrl] <String>]
	[-SkipUpload]
	[-SkipSubscription]
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
Create a new microservice or upload a new microservice binary to an already running microservice.
By default the microservice will
also be subscribed to/enabled.

The zip file needs to follow the Cumulocity Microservice format.

This cmdlet has several operations

## EXAMPLES

### EXAMPLE 1
```
New-Microservice -File "myapp.zip"
```

Upload microservice binary.
The name of the microservice will be named after the zip file name (without the extension)

If the microservice already exists, then the only the microservice binary will be updated.

### EXAMPLE 2
```
New-Microservice -Name "myapp" -File "myapp.zip"
```

Upload microservice binary with a custom name.
Note: If the microservice already exists in the platform

### EXAMPLE 3
```
New-Microservice -Name "myapp" -File "./cumulocity.json" -SkipUpload
```

Create a microservice placeholder named "myapp" for use for local development of a microservice.

The `-File` parameter is provided with the microserivce's manifest file `cumulocity.json` to set the correct required roles of the bootstrap
user which will be automatically created by Cumulocity.

The microservice's bootstrap credentials can be retrieved using `Get-MicroserviceBootstrapUser` cmdlet.

This example is usefuly for local development only, when you want to run the microservice locally (not hosted in Cumulocity).

## PARAMETERS

### -File
File to be uploaded as a binary (required)

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

### -Name
Name of the microservice.
An id is also accepted however the name have been previously uploaded.

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

### -Key
Shared secret of application.
Defaults to application name if not provided.

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

### -Availability
Access level for other tenants. 
Possible values are : MARKET, PRIVATE (default)

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

### -ContextPath
ContextPath of the hosted application.
Required when application type is HOSTED

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

### -ResourcesUrl
URL to application base directory hosted on an external server.
Required when application type is HOSTED

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

### -SkipUpload
Skip the uploading of the microservice binary.
This is helpful if you want to run the microservice locally
and you only need the microservice place holder in order to create microservice bootstrap credentials.

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

### -SkipSubscription
Don't subscribe to the microservice after it has been created and uploaded

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
Position: 7
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
Position: 8
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
Position: 9
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
This cmdlet does not support template variables

## RELATED LINKS
