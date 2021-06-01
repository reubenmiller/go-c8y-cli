# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-GroupFromGroup {
<#
.SYNOPSIS
Delete child group reference

.DESCRIPTION
Delete child group reference

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicegroups_unassignGroup

.EXAMPLE
PS> Remove-GroupFromGroup -Id $Group.id -Child $ChildGroup.id

Unassign a child group from its parent


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device group (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Child device group (required)
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicegroups unassignGroup"
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
            | c8y devicegroups unassignGroup $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Child `
            | Group-ClientRequests `
            | c8y devicegroups unassignGroup $c8yargs
        }
        
    }

    End {}
}
