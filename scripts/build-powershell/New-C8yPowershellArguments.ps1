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
        "[]agent" { "object[]"; break }
        "[]device" { "object[]"; break }
        "[]devicegroup" { "object[]"; break }
        "[]deviceprofile" { "object[]"; break }
        "[]id" { "object[]"; break }
        "[]smartgroup" { "object[]"; break }
        "[]role" { "object[]"; break }
        "[]roleself" { "object[]"; break }
        "[]string" { "string[]"; break }
        "[]configuration" { "object[]"; break }
        "[]software" { "object[]"; break }
        "softwareName" { "object[]"; break }
        "[]softwareversion" { "object[]"; break }
        "softwareversionName" { "object[]"; break }
        "[]firmware" { "object[]"; break }
        "firmwareName" { "object[]"; break }
        "[]firmwareversion" { "object[]"; break }
        "firmwareversionName" { "object[]"; break }
        "[]firmwarepatch" { "object[]"; break }
        "firmwarepatchName" { "object[]"; break }
        "[]stringcsv" { "string[]"; break }
        "[]tenant" { "object[]"; break }
        "[]user" { "object[]"; break }
        "[]usergroup" { "object[]"; break }
        "[]userself" { "object[]"; break }
        "application" { "object[]"; break }
        "applicationname" { "string"; break }
        "hostedapplication" { "object[]"; break }
        "boolean" { "switch"; break }
        "booleanDefault" { "switch"; break }
        "optional_fragment" { "switch"; break }
        "datefrom" { "string"; break }
        "datetime" { "string"; break }
        "date" { "string"; break }
        "dateto" { "string"; break }
        "directory" { "string"; break }
        "file" { "string"; break }
        "float" { "float"; break }
        "fileContents" { "string"; break }
        "attachment" { "string"; break }
        "binaryUploadURL" { "string"; break }
        "id" { "object[]"; break }
        "integer" { "long"; break }
        "json" { "object"; break }
        "json_custom" { "object"; break }
        "microservice" { "object[]"; break }
        "microserviceinstance" { "string"; break }
        "set" { "object[]"; break }
        "source" { "object"; break }
        "string" { "string"; break }
        "[]devicerequest" { "object[]"; break }
        "strings" { "string"; break }
        "tenant" { "object"; break }
        "tenantname" { "string"; break }
        "[]certificate" { "object[]"; break }
        "certificatefile" { "string"; break }

        # stringStatic
        "stringStatic" { $Ignore = $true; ""; break }

        # queryExpression
        "queryExpression" { $Ignore = $true; ""; break }

        # Complex lookup types. These should not be visible in powershell
        "softwareDetails" { $Ignore = $true; ""; break }
        "firmwareDetails" { $Ignore = $true; ""; break }
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
