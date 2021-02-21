# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ManagedObjectNative {
    <#
    .SYNOPSIS
    Get managed objects/s

    .DESCRIPTION
    Get a managed object by id

    .EXAMPLE
    PS> Get-ManagedObject -Id $mo.id

    Get a managed object

    .EXAMPLE
    PS> Get-ManagedObject -Id $mo.id | Get-ManagedObject

    Get a managed object (using pipeline)

    .EXAMPLE
    PS> Get-ManagedObject -Id $mo.id -WithParents

    Get a managed object with parent references


    #>
    [cmdletbinding(SupportsShouldProcess = $true,
        PositionalBinding = $true,
        HelpUri = '',
        ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # ManagedObject id (required)
        [Parameter(Mandatory = $true,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true)]
        [object[]]
        $Id,

        # include a flat list of all parents and grandparents of the given object
        [Parameter()]
        [switch]
        $WithParents
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get" -BoundParameters $PSBoundParameters
    }

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("WithParents")) {
            $Parameters["withParents"] = $WithParents
        }

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }
        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "inventory get"
        $OutputOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        $Id `
        | c8y inventory get $c8yargs
    }

    End {}
}
