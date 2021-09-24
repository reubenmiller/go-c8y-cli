# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-DeviceProfile {
<#
.SYNOPSIS
Update device profile

.DESCRIPTION
Update an existing device profile (managedObject)

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/deviceprofiles_update

.EXAMPLE
PS> Update-DeviceProfile -Id $mo.id -Data @{ com_my_props = @{ value = 1 } }

Update a device profile

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Update-DeviceProfile -Data @{ com_my_props = @{ value = 1 } }

Update a device profile (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device profile (managedObject) id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # New device profile name
        [Parameter()]
        [string]
        $NewName,

        # Device type filter. Only allow device profile to be applied to devices of this type
        [Parameter()]
        [string]
        $DeviceType
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "deviceprofiles update"
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
            | c8y deviceprofiles update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y deviceprofiles update $c8yargs
        }
        
    }

    End {}
}
