# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-Firmware {
<#
.SYNOPSIS
Update firmware

.DESCRIPTION
Update an existing firmware package (managedObject)

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/firmware_update

.EXAMPLE
PS> Update-Firmware -Id $mo.id -Data @{ com_my_props = @{ value = 1 } }

Update a firmware package

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Update-Firmware -Data @{ com_my_props = @{ value = 1 } }

Update a firmware package (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Firmware package (managedObject) id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # New firmware package name
        [Parameter()]
        [string]
        $NewName,

        # Description of the firmware package
        [Parameter()]
        [string]
        $Description,

        # Device type filter. Only allow firmware to be applied to devices of this type
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "firmware update"
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
            | c8y firmware update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y firmware update $c8yargs
        }
        
    }

    End {}
}
