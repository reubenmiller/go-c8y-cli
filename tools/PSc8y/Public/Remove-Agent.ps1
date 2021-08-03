# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-Agent {
<#
.SYNOPSIS
Delete agent

.DESCRIPTION
Delete an agent from the platform. This will delete all data associated to the agent
(i.e. alarms, events, operations and measurements)


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/agents_delete

.EXAMPLE
PS> Remove-Agent -Id $agent.id

Remove agent by id

.EXAMPLE
PS> Remove-Agent -Id $agent.name

Remove agent by name

.EXAMPLE
PS> Remove-Agent -Id "agent01" -WithDeviceUser

Delete agent and related device user/credentials


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Delete associated device owner
        [Parameter()]
        [switch]
        $WithDeviceUser,

        # Agent ID (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Remove all child devices and child assets will be deleted recursively. By default, the delete operation is propagated to the subgroups only if the deleted object is a group
        [Parameter()]
        [switch]
        $Cascade
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "agents delete"
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
            | c8y agents delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y agents delete $c8yargs
        }
        
    }

    End {}
}
