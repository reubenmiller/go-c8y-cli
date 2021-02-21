Function Update-DeviceNative {
    <#
    .SYNOPSIS
    Update device

    .DESCRIPTION
    Update properties of an existing device

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
    [cmdletbinding(SupportsShouldProcess = $true,
        PositionalBinding = $true,
        HelpUri = '',
        ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device ID (required)
        [Parameter(Mandatory = $true,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true)]
        [object[]]
        $Id,

        # Device name
        [Parameter()]
        [string]
        $NewName
    )
    DynamicParam {
        Get-ClientCommonParameters -Type @("Update", "Template") -BoundParameters $PSBoundParameters
    }

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("NewName")) {
            $Parameters["newName"] = $NewName
        }

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }
        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices update"
        $OutputOptions = @{
            Type            = "application/vnd.com.nsn.cumulocity.customDevice+json"
            ItemType        = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        if (!$Force -and !$WhatIfPreference) {
            $items = Expand-Id -InputObject $Id

            $shouldContinue = $PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $items)
            )
            if (!$shouldContinue) {
                return
            }
        }

        $Id `
        | c8y devices update $c8yargs `
        | ConvertFrom-ClientOutput @OutputOptions
    }

    End {}
}
