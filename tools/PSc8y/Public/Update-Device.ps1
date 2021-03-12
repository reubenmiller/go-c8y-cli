# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-Device {
<#
.SYNOPSIS
Update device

.DESCRIPTION
Update properties of an existing device

.LINK
c8y devices update

.EXAMPLE
PS> Update-Device -Id $device.id -NewName "MyNewName"

Update device by id

.EXAMPLE
PS> Update-Device -Id $device.name -NewName "MyNewName"

Update device by name

.EXAMPLE
PS> Update-Device -Id $device.name -Data @{ "myValue" = @{ value1 = $true } }

Update device custom properties


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device ID (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Device name
        [Parameter()]
        [string]
        $NewName
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customDevice+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y devices update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y devices update $c8yargs
        }
        
    }

    End {}
}
