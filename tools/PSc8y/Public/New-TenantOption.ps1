# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-TenantOption {
<#
.SYNOPSIS
Create tenant option

.DESCRIPTION
Create tenant option

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/tenantoptions_create

.EXAMPLE
PS> New-TenantOption -Category "c8y_cli_tests" -Key "$option1" -Value "1"

Create a tenant option


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Category of option
        [Parameter()]
        [string]
        $Category,

        # Key of option
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Key,

        # Value of option
        [Parameter()]
        [string]
        $Value
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenantoptions create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.option+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Key `
            | Group-ClientRequests `
            | c8y tenantoptions create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Key `
            | Group-ClientRequests `
            | c8y tenantoptions create $c8yargs
        }
        
    }

    End {}
}
