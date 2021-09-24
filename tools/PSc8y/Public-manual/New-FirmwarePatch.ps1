Function New-FirmwarePatch {
<#
.SYNOPSIS
Create firmware package version patch

.DESCRIPTION
Create a new firmware package (managedObject)

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/firmware_patches_create

.EXAMPLE
PS> New-FirmwarePatch -Firmware "UBUNTU_20_04" -Version "20.4.1" -DependencyVersion "20.4.0" -Url "https://example.com/binary/12345

Create a new patch (with external URL) to an existing firmware version


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
        $Firmware,

        # Patch version, i.e. 1.0.0
        [Parameter()]
        [string]
        $Version,

        # URL to the firmware patch
        [Parameter()]
        [string]
        $Url,

        # File to be uploaded
        [Parameter()]
        [string]
        $File,

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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "firmware patches create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Firmware `
            | Group-ClientRequests `
            | c8y firmware patches create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Firmware `
            | Group-ClientRequests `
            | c8y firmware patches create $c8yargs
        }
        
    }

    End {}
}
