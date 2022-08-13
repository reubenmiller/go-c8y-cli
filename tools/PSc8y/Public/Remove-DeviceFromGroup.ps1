# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-DeviceFromGroup {
<#
.SYNOPSIS
Unassign device from group

.DESCRIPTION
Unassign/delete a device from a group

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicegroups_devices_unassign

.EXAMPLE
PS> Remove-DeviceFromGroup -Group $Group.id -Child $ChildDevice.id

Unassign a child device from its parent asset


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device group (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Group,

        # Child device (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Child
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicegroups devices unassign"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Child `
            | Group-ClientRequests `
            | c8y devicegroups devices unassign $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Child `
            | Group-ClientRequests `
            | c8y devicegroups devices unassign $c8yargs
        }
        
    }

    End {}
}
