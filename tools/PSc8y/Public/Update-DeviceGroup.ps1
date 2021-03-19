# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-DeviceGroup {
<#
.SYNOPSIS
Update device group

.DESCRIPTION
Update properties of an existing device group, for example name or any other custom properties.


.LINK
c8y devicegroups update

.EXAMPLE
PS> Update-DeviceGroup -Id $group.id -Name "MyNewName"

Update device group by id

.EXAMPLE
PS> Update-DeviceGroup -Id $group.name -Name "MyNewName"

Update device group by name

.EXAMPLE
PS> Update-DeviceGroup -Id $group.name -Data @{ "myValue" = @{ value1 = $true } }

Update device group custom properties


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device group ID (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Device group name
        [Parameter()]
        [string]
        $Name
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicegroups update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customDeviceGroup+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y devicegroups update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devicegroups update $c8yargs
        }
        
    }

    End {}
}
