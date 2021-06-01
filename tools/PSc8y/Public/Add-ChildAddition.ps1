# Code generated from specification version 1.0.0: DO NOT EDIT
Function Add-ChildAddition {
<#
.SYNOPSIS
Assign child addition

.DESCRIPTION
Add an existing managed object as a child addition to another existing managed object

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/inventory_additions_assign

.EXAMPLE
PS> Add-ChildAddition -Id $software.id -NewChild $version.id

Add a related managed object as a child to an existing managed object


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

        # New managed object that will be added as a child addition (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $NewChild
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory additions assign"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReference+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $NewChild `
            | Group-ClientRequests `
            | c8y inventory additions assign $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $NewChild `
            | Group-ClientRequests `
            | c8y inventory additions assign $c8yargs
        }
        
    }

    End {}
}
