# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-DeviceProfile {
<#
.SYNOPSIS
Create device profile

.DESCRIPTION
Create a new device profile (managedObject)

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/deviceprofiles_create

.EXAMPLE
PS> New-ManagedObject -Name "python3-requests" -Description "python requests library" -Data @{$type=@{}}

Create a managed object


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # name
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Name,

        # Device type filter. Only allow device profile to be applied to devices of this type
        [Parameter()]
        [string]
        $DeviceType
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "deviceprofiles create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Name `
            | Group-ClientRequests `
            | c8y deviceprofiles create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Name `
            | Group-ClientRequests `
            | c8y deviceprofiles create $c8yargs
        }
        
    }

    End {}
}
