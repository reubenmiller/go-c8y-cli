# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-SmartGroup {
<#
.SYNOPSIS
Delete smart group

.DESCRIPTION
Delete an existing smart group by id or name. Deleting a smart group will not affect any of the devices related to it.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/smartgroups_delete

.EXAMPLE
PS> Remove-SmartGroup -Id $smartgroup.id

Remove smart group by id

.EXAMPLE
PS> Remove-SmartGroup -Id $smartgroup.name

Remove smart group by name


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Smart group ID (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "smartgroups delete"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y smartgroups delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y smartgroups delete $c8yargs
        }
        
    }

    End {}
}
