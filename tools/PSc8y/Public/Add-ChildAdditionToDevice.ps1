# Code generated from specification version 1.0.0: DO NOT EDIT
Function Add-ChildAdditionToDevice {
<#
.SYNOPSIS
Assign child addition

.DESCRIPTION
Create a child addition reference to an existing device

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devices_additions_assign

.EXAMPLE
PS> Add-ChildAdditionToDevice -Device $Device.id -Child $ChildDevice.id

Assign a managed object as a child addition to an existing device

.EXAMPLE
PS> Get-ManagedObject -Id $ChildDevice.id | Add-ChildDeviceToDevice -Device $Device.id

Assign a managed object as a child addition to an existing device (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device. (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Device,

        # New child addition (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Child
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices additions assign"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReference+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Child `
            | Group-ClientRequests `
            | c8y devices additions assign $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Child `
            | Group-ClientRequests `
            | c8y devices additions assign $c8yargs
        }
        
    }

    End {}
}
