# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-DeviceChild {
<#
.SYNOPSIS
Create child

.DESCRIPTION
Create a new managed object and assign it to an existing managed object as a child

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_children_create

.EXAMPLE
PS> New-DeviceChild -Id $software.id -Data "custom.value=test" -Global -ChildType addition

Create a child addition and link it to an existing managed object


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Managed object id where the child addition will be added to (required)
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

        # Enable global access to the managed object
        [Parameter()]
        [switch]
        $Global
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices children create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y devices children create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devices children create $c8yargs
        }
        
    }

    End {}
}
