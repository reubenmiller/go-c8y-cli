# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-Software {
<#
.SYNOPSIS
Update software

.DESCRIPTION
Update an existing software package (managedObject)

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/software_update

.EXAMPLE
PS> Update-Software -Id $mo.id -Data @{ com_my_props = @{ value = 1 } }

Update a software package

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Update-Software -Data @{ com_my_props = @{ value = 1 } }

Update a software package (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Software package (managedObject) id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # New software package name
        [Parameter()]
        [string]
        $NewName,

        # Description of the software package
        [Parameter()]
        [string]
        $Description,

        # Device type filter. Only allow software to be applied to devices of this type
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "software update"
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
            | c8y software update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y software update $c8yargs
        }
        
    }

    End {}
}
