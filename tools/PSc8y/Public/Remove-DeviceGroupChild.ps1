# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-DeviceGroupChild {
<#
.SYNOPSIS
Unassign child

.DESCRIPTION
Unassign/delete an managed object as a child to an existing managed object

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicegroups_children_unassign

.EXAMPLE
PS> Remove-DeviceGroupChild -Id $software.id -Child $version.id -ChildType addition

Unassign a child addition from its parent managed object


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Managed object id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Child relationship type (required)
        [Parameter(Mandatory = $true)]
        [ValidateSet('addition','asset','device')]
        [string]
        $ChildType,

        # Child managed object id (required)
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicegroups children unassign"
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
            | c8y devicegroups children unassign $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Child `
            | Group-ClientRequests `
            | c8y devicegroups children unassign $c8yargs
        }
        
    }

    End {}
}
