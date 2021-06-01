# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-SmartGroup {
<#
.SYNOPSIS
Update smart group

.DESCRIPTION
Update properties of an existing smart group

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/smartgroups_update

.EXAMPLE
PS> Update-SmartGroup -Id $smartgroup.id -NewName "MyNewName"

Update smart group by id

.EXAMPLE
PS> Update-SmartGroup -Id $smartgroup.name -NewName "MyNewName"

Update smart group by name

.EXAMPLE
PS> Update-SmartGroup -Id $smartgroup.name -Data @{ "myValue" = @{ value1 = $true } }

Update smart group custom properties


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
        $Id,

        # New smart group name
        [Parameter()]
        [string]
        $NewName,

        # New query
        [Parameter()]
        [string]
        $Query
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "smartgroups update"
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
            | c8y smartgroups update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y smartgroups update $c8yargs
        }
        
    }

    End {}
}
