# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-DeviceProfile {
<#
.SYNOPSIS
Get device profile

.DESCRIPTION
Get an existing device profile (managedObject)

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/deviceprofiles_get

.EXAMPLE
PS> Get-DeviceProfile -Id $mo.id

Get a device profile

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Get-DeviceProfile

Get a device profile (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # DeviceProfile (managedObject) id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        []
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "deviceprofiles get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y deviceprofiles get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y deviceprofiles get $c8yargs
        }
        
    }

    End {}
}
