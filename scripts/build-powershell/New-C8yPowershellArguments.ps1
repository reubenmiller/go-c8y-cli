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

    # Type Definition
    $DataType = switch ($Type) {
        "[]agent" { "object[]" }
        "[]device" { "object[]" }
        "[]devicegroup" { "object[]" }
        "[]role" { "object[]" }
        "[]roleself" { "object[]" }
        "[]string" { "string[]" }
        "[]stringcsv" { "string[]" }
        "[]tenant" { "object[]" }
        "[]user" { "object[]" }
        "[]usergroup" { "object[]" }
        "[]userself" { "object[]" }
        "application" { "object[]" }
        "boolean" { "switch" }
        "datefrom" { "string" }
        "datetime" { "string" }
        "dateto" { "string" }
        "directory" { "string" }
        "file" { "string" }
        "float" { "float" }
        "fileContents" { "string" }
        "attachment" { "string" }
        "id" { "object" }
        "integer" { "long" }
        "json" { "object" }
        "json_custom" { "object" }
        "microservice" { "object[]" }
        "set" { "object[]" }
        "source" { "object" }
        "string" { "string" }
        "strings" { "string" }
        "tenant" { "object" }
        default {
            Write-Error "Unsupported Type. $_"
        }
    }

    New-Object psobject -Property @{
        Name = $NameLocalVariable
        Type = $DataType
        Definition = $ParameterDefinition
        Description = "$Description"

    }
}
