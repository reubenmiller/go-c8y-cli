# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-ChildDeviceFromDevice {
<#
.SYNOPSIS
Delete child device reference

.DESCRIPTION
Delete child device reference

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_children_unassign

.EXAMPLE
PS> Remove-ChildDeviceFromDevice -Device $Device.id -Child $ChildDevice.id

Unassign a child device from its parent device


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # ManagedObject id (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Device,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices children unassign"
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
            | c8y devices children unassign $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Child `
            | Group-ClientRequests `
            | c8y devices children unassign $c8yargs
        }
        
    }

    End {}
}
