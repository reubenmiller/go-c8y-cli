﻿Function New-TestDeviceGroup {
<# 
.SYNOPSIS
Create a new test device group

.DESCRIPTION
Create a new test device group with a randomized name. Useful when performing mockups or prototyping.

.EXAMPLE
New-TestDeviceGroup

Create a test device group

.EXAMPLE
1..10 | Foreach-Object { New-TestDeviceGroup -Force }

Create 10 test device groups all with unique names

.EXAMPLE
New-TestDeviceGroup -TotalDevices 10

Create a test device group with 10 newly created devices
#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "High"
    )]
    Param(
        # Device group name prefix which is added before the randomized string
        [Parameter(
            Mandatory = $false,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
        [string] $Name = "testgroup",

        # Group type. Only device groups of type `Group` are visible as root folders in the UI
        [ValidateSet("Group", "SubGroup")]
        [string] $Type = "Group",

        # Number of devices to create and assign to the group
        [int]
        $TotalDevices = 0,

        # Cumulocity processing mode
        [Parameter()]
        [AllowNull()]
        [AllowEmptyString()]
        [ValidateSet("PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP")]
        [string]
        $ProcessingMode,

        # Template (jsonnet) file to use to create the request body.
        [Parameter()]
        [string]
        $Template,

        # Variables to be used when evaluating the Template. Accepts json or json shorthand, i.e. "name=peter"
        [Parameter()]
        [string]
        $TemplateVars,

        # Don't prompt for confirmation
        [switch] $Force
    )

    Process {
        $Data = @{
            c8y_IsDeviceGroup = @{ }
        }

        switch ($Type) {
            "SubGroup" {
                $Data.type = "c8y_DeviceSubGroup"
                break;
            }
            default {
                $Data.type = "c8y_DeviceGroup"
                break;
            }
        }

        $GroupName = New-RandomString -Prefix "${Name}_"
        $Group = PSc8y\New-ManagedObject `
            -Name $GroupName `
            -Data $Data `
            -ProcessingMode:$ProcessingMode `
            -Template:$Template `
            -TemplateVars:$TemplateVars `
            -Force:$Force
        
        if ($TotalDevices -gt 0) {
            for ($i = 0; $i -lt $TotalDevices; $i++) {
                $iDevice = PSc8y\New-TestAgent -Force
                $null = PSc8y\Add-AssetToGroup -Group $Group.id -NewChildDevice $iDevice.id -Force
            }
        }

        Write-Output $Group
    }
}
