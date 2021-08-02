# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-DeviceProfile {
<#
.SYNOPSIS
Delete device profile

.DESCRIPTION
Delete an existing device profile

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/deviceprofiles_delete

.EXAMPLE
PS> Remove-DeviceProfile -Id $mo.id

Delete a device profile

.EXAMPLE
PS> Get-ManagedObject -Id $mo.id | Remove-DeviceProfile

Delete a device profile (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # DeviceProfile Package (managedObject) id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        []
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "deviceprofiles delete"
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
            | c8y deviceprofiles delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y deviceprofiles delete $c8yargs
        }
        
    }

    End {}
}
