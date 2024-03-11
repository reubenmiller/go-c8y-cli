Function New-C8yPowershellArguments {
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [string] $Name,

        [Parameter(
            Mandatory = $true,
            Position = 1
        )]
        [string] $Type,

        [string] $Required,

        [string] $OptionName,

        [string] $Description,

        [string] $Default,

        [switch] $ReadFromPipeline
    )

    # Format variable name
    $NameLocalVariable = $Name[0].ToString().ToUpperInvariant() + $Name.Substring(1)

    $ParameterDefinition = New-Object System.Collections.ArrayList

    if ($Required -match "true|yes") {
        $null = $ParameterDefinition.Add("Mandatory = `$true")
        $Description = "${Description} (required)"
    }

    # TODO: Do we need to add Position = x? to the ParameterDefinition

    # Add alias
    if ($UseOption) {
        $null = $ParameterDefinition.Add("Alias = `"$OptionName`"")
    }

    # Add Piped argument
    if ($Type -match "(source|id)" -or $ReadFromPipeline) {
        $null = $ParameterDefinition.Add("ValueFromPipeline=`$true")
        $null = $ParameterDefinition.Add("ValueFromPipelineByPropertyName=`$true")
    }

    $Ignore = $false

    # Type Definition
    $DataType = switch ($Type) {
        "agent[]" { "object[]"; break }
        "certificate[]" { "object[]"; break }
        "configuration[]" { "object[]"; break }
        "device[]" { "object[]"; break }
        "devicegroup[]" { "object[]"; break }
        "deviceprofile[]" { "object[]"; break }
        "devicerequest[]" { "object[]"; break }
        "firmware[]" { "object[]"; break }
        "firmwarepatch[]" { "object[]"; break }
        "firmwareversion[]" { "object[]"; break }
        "id[]" { "object[]"; break }
        "role[]" { "object[]"; break }
        "roleself[]" { "object[]"; break }
        "smartgroup[]" { "object[]"; break }
        "software[]" { "object[]"; break }
        "softwareversion[]" { "object[]"; break }
        "deviceservice[]" { "object[]"; break }
        "string[]" { "string[]"; break }
        "stringcsv[]" { "string[]"; break }
        "[]tenant" { "object[]"; break }
        "user[]" { "object[]"; break }
        "usergroup[]" { "object[]"; break }
        "userself[]" { "object[]"; break }
        "application" { "object[]"; break }
        "application_with_versions" { "object[]"; break }
        "applicationname" { "string"; break }
        "attachment" { "string"; break }
        "binaryUploadURL" { "string"; break }
        "boolean" { "switch"; break }
        "booleanDefault" { "switch"; break }
        "certificatefile" { "string"; break }
        "date" { "string"; break }
        "datefrom" { "string"; break }
        "datetime" { "string"; break }
        "dateto" { "string"; break }
        "directory" { "string"; break }
        "file" { "string"; break }
        "formDataFile" { "string"; break }
        "fileContents" { "string"; break }
        "firmwareName" { "object[]"; break }
        "firmwarepatchName" { "object[]"; break }
        "firmwareversionName" { "object[]"; break }
        "float" { "float"; break }
        "hostedapplication" { "object[]"; break }
        "id" { "object[]"; break }
        "integer" { "long"; break }
        "inventoryChildType" { "string"; break }
        "json_custom" { "object"; break }
        "json" { "object"; break }
        "microservice" { "object[]"; break }
        "microserviceinstance" { "string"; break }
        "microservicename" { "object[]"; break }
        "optional_fragment" { "switch"; break }
        "set" { "object[]"; break }
        "softwareName" { "object[]"; break }
        "softwareversionName" { "object[]"; break }
        "source" { "object"; break }
        "string" { "string"; break }
        "strings" { "string"; break }
        "subscriptionId" { "string"; break }
        "subscriptionName" { "string"; break }
        "tenant" { "object"; break }
        "tenantname" { "string"; break }

        # stringStatic
        "stringStatic" { $Ignore = $true; ""; break }

        # queryExpression
        "queryExpression" { $Ignore = $true; ""; break }

        # Complex lookup types. These should not be visible in powershell
        "softwareDetails" { $Ignore = $true; ""; break }
        "firmwareDetails" { $Ignore = $true; ""; break }
        "configurationDetails" { $Ignore = $true; ""; break }
        default {
            Write-Error "Unsupported Type. $_"
        }
    }

    New-Object psobject -Property @{
        Name = $NameLocalVariable
        Type = $DataType
        Definition = $ParameterDefinition
        Description = "$Description"
        Ignore = $Ignore
    }
}
