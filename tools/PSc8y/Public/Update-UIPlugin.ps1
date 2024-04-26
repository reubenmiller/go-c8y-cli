# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-UIPlugin {
<#
.SYNOPSIS
Update UI plugin details

.DESCRIPTION
Update details of an existing UI plugin

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/ui_plugins_update

.EXAMPLE
PS> Update-UIPlugin -Id $App.name -Availability "MARKET"

Update plugin availability to MARKET


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Plugin (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Name of the plugin
        [Parameter()]
        [string]
        $Name,

        # Shared secret of the plugin
        [Parameter()]
        [string]
        $Key,

        # Access level for other tenants
        [Parameter()]
        [ValidateSet('SHARED','PRIVATE','MARKET')]
        [string]
        $Availability,

        # contextPath of the plugin
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "ui plugins update"
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
            | c8y ui plugins update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y ui plugins update $c8yargs
        }
        
    }

    End {}
}
