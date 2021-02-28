# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Device {
<#
.SYNOPSIS
Create a device

.DESCRIPTION
Create a device (managed object) with the special c8y_IsDevice fragment.


.EXAMPLE
PS> New-Device -Name $DeviceName

Create device

.EXAMPLE
PS> New-Device -Name $DeviceName -Data @{ myValue = @{ value1 = $true } }

Create device with custom properties

.EXAMPLE
PS> New-Device -Template "{ name: '$DeviceName' }"


Create device using a template


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device name
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Name,

        # Device type
        [Parameter()]
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customDevice+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        $Force = if ($PSBoundParameters.ContainsKey("Force")) { $PSBoundParameters["Force"] } else { $False }
        if (!$Force -and !$WhatIfPreference) {
            $items = (PSc8y\Expand-Id $Name)

            $shouldContinue = $PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $items)
            )
            if (!$shouldContinue) {
                return
            }
        }

        if ($ClientOptions.ConvertToPS) {
            $Name `
            | Group-ClientRequests `
            | c8y devices create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Name `
            | Group-ClientRequests `
            | c8y devices create $c8yargs
        }
        
    }

    End {}
}
