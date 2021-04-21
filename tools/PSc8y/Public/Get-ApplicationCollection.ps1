# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ApplicationCollection {
<#
.SYNOPSIS
Get application collection

.DESCRIPTION
Get a collection of applications by a given filter

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/applications_list

.EXAMPLE
PS> Get-ApplicationCollection -PageSize 100

Get applications


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Application type
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [ValidateSet('APAMA_CEP_RULE','EXTERNAL','HOSTED','MICROSERVICE')]
        [object[]]
        $Type
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "applications list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.applicationCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.application+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Type `
            | Group-ClientRequests `
            | c8y applications list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Type `
            | Group-ClientRequests `
            | c8y applications list $c8yargs
        }
        
    }

    End {}
}
