# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-DeviceGroup {
<#
.SYNOPSIS
Update device group

.DESCRIPTION
Update properties of an existing device group, for example name or any other custom properties.


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
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
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
        $Name,

        # Additional properties of the inventory.
        [Parameter()]
        [object]
        $Data,

        # Show the full (raw) response from Cumulocity including pagination information
        [Parameter()]
        [switch]
        $Raw,

        # Write the response to file
        [Parameter()]
        [string]
        $OutputFile,

        # Ignore any proxy settings when running the cmdlet
        [Parameter()]
        [switch]
        $NoProxy,

        # Specifiy alternative Cumulocity session to use when running the cmdlet
        [Parameter()]
        [string]
        $Session,

        # TimeoutSec timeout in seconds before a request will be aborted
        [Parameter()]
        [double]
        $TimeoutSec,

        # Don't prompt for confirmation
        [Parameter()]
        [switch]
        $Force
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Name")) {
            $Parameters["name"] = $Name
        }
        if ($PSBoundParameters.ContainsKey("Data")) {
            $Parameters["data"] = ConvertTo-JsonArgument $Data
        }
        if ($PSBoundParameters.ContainsKey("OutputFile")) {
            $Parameters["outputFile"] = $OutputFile
        }
        if ($PSBoundParameters.ContainsKey("NoProxy")) {
            $Parameters["noProxy"] = $NoProxy
        }
        if ($PSBoundParameters.ContainsKey("Session")) {
            $Parameters["session"] = $Session
        }
        if ($PSBoundParameters.ContainsKey("TimeoutSec")) {
            $Parameters["timeout"] = $TimeoutSec * 1000
        }

    }

    Process {
        foreach ($item in (PSc8y\Expand-Id $Id)) {
            if ($item) {
                $Parameters["id"] = if ($item.id) { $item.id } else { $item }
            }

            if (!$Force -and
                !$WhatIfPreference -and
                !$PSCmdlet.ShouldProcess(
                    (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                    (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
                )) {
                continue
            }

            Invoke-ClientCommand `
                -Noun "devices" `
                -Verb "updateGroup" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.customDeviceGroup+json" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }

    End {}
}
