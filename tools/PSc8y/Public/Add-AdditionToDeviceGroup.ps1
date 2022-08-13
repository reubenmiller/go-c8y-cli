# Code generated from specification version 1.0.0: DO NOT EDIT
Function Add-AdditionToDeviceGroup {
<#
.SYNOPSIS
Assign asset to group

.DESCRIPTION
Assigns an asset to a group. The device will be a childAddition of the group

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicegroups_additions_assign

.EXAMPLE
PS> Add-AdditionToDeviceGroup -Group $Group.id -Child $Device.id

Add a device to a group

.EXAMPLE
PS> Add-AdditionToDeviceGroup -Group $Group -Child $Device

Add a device to a group by passing device and groups instead of an id or name

.EXAMPLE
PS> Get-Device $Device1.name, $Device2.name | Add-AdditionToDeviceGroup -Group $Group.id

Add multiple devices to a group. Alternatively `Get-DeviceCollection` can be used
to filter for a collection of devices and assign the results to a single group.



#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Group (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Group,

        # New device to be added to the group as an child asset (required)
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicegroups additions assign"
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
            | c8y devicegroups additions assign $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Child `
            | Group-ClientRequests `
            | c8y devicegroups additions assign $c8yargs
        }
        
    }

    End {}
}
