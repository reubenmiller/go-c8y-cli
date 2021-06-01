# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-DeviceGroup {
<#
.SYNOPSIS
Create device group

.DESCRIPTION
Create a new device group to logically group one or more devices


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/devicegroups_create

.EXAMPLE
PS> New-DeviceGroup -Name $GroupName

Create device group

.EXAMPLE
PS> New-DeviceGroup -Name $GroupName -Data @{ "myValue" = @{ value1 = $true } }

Create device group with custom properties


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device group name
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Name,

        # Device group type (c8y_DeviceGroup (root folder) or c8y_DeviceSubGroup (sub folder)). Defaults to c8y_DeviceGroup
        [Parameter()]
        [ValidateSet('c8y_DeviceGroup','c8y_DeviceSubGroup')]
        [string]
        $Type
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicegroups create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customDeviceGroup+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Name `
            | Group-ClientRequests `
            | c8y devicegroups create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Name `
            | Group-ClientRequests `
            | c8y devicegroups create $c8yargs
        }
        
    }

    End {}
}
