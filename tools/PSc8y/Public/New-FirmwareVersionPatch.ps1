﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-FirmwareVersionPatch {
<#
.SYNOPSIS
Create firmware package version patch

.DESCRIPTION
Create a new firmware package (managedObject)

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/firmware_versions_createPatch

.EXAMPLE
PS> New-ManagedObject -Name "python3-requests" -Description "python requests library" -Data @{$type=@{}}

Create a new version to an existing firmware package


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Firmware package id where the version will be added to
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $FirmwareId,

        # Patch version, i.e. 1.0.0
        [Parameter()]
        [string]
        $Version,

        # URL to the firmware patch
        [Parameter()]
        [string]
        $Url,

        # Existing firmware version that the patch is dependent on
        [Parameter()]
        [string]
        $DependencyVersion
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "firmware versions createPatch"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $FirmwareId `
            | Group-ClientRequests `
            | c8y firmware versions createPatch $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $FirmwareId `
            | Group-ClientRequests `
            | c8y firmware versions createPatch $c8yargs
        }
        
    }

    End {}
}