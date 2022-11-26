---
category: Sessions
external help file: PSc8y-help.xml
id: New-Session
Module Name: PSc8y
online version:
schema: 2.0.0
slug: /docs/cli/psc8y/Sessions/new-session
title: New-Session
---



## SYNOPSIS
Create a new Cumulocity Session

## SYNTAX

```
New-Session
	[-Host] <String>
	[[-Tenant] <String>]
	[[-Username] <Object>]
	[[-Password] <Object>]
	[[-Name] <String>]
	[[-Description] <String>]
	[-NoTenantPrefix]
	[-AllowInsecure]
	[<CommonParameters>]
```

## DESCRIPTION
Create a new Cumulocity session which can be used by the cmdlets

## EXAMPLES

### EXAMPLE 1
```
New-Session -Name "develop" -Host "my-tenant-name.eu-latest.cumulocity.com"
```

Create a new Cumulocity session called develop

### EXAMPLE 2
```
New-Session -Host "my-tenant-name.eu-latest.cumulocity.com"
```

Create a new Cumulocity session.
It will prompt for the username and password.

## PARAMETERS

### -Host
Host url, i.e.
https://my-tenant-name.eu-latest.cumulocity.com

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

### -Tenant
Tenant id, i.e.
t12345

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

### -Username
Username

```yaml
Type: Object
Parameter Sets: (All)
Aliases:

Required: False
Position: 3
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Password
Password

```yaml
Type: Object
Parameter Sets: (All)
Aliases:

Required: False
Position: 4
Default value: None
Accept pipeline input: False
Accept wildcard characters: False
```

### -Name
Name of the Cumulocity session

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

### -Description
Description

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

### -NoTenantPrefix
Don't use tenant name as a prefix to the user name when using Basic Authentication

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

### -AllowInsecure
Allow insecure connection (e.g.
when using self-signed certificates)

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

### None
## NOTES

## RELATED LINKS

[https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/sessions_create](https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/sessions_create)

