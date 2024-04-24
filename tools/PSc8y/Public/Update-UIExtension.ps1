# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-UIExtension {
<#
.SYNOPSIS
Update UI extension details

.DESCRIPTION
Update details of an existing UI extension

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/ui_extensions_update

.EXAMPLE
PS> Update-UIExtension -Id $App.name -Availability "MARKET"

Update application availability to MARKET


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Extension id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Name of the extension
        [Parameter()]
        [string]
        $Name,

        # Shared secret of the extension
        [Parameter()]
        [string]
        $Key,

        # Access level for other tenants. Possible values are : MARKET, PRIVATE (default)
        [Parameter()]
        [ValidateSet('SHARED','PRIVATE','MARKET')]
        [string]
        $Availability,

        # contextPath of the extension
        [Parameter()]
        [string]
        $ContextPath
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "ui extensions update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.application+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y ui extensions update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y ui extensions update $c8yargs
        }
        
    }

    End {}
}
