# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-EventCollection {
<#
.SYNOPSIS
Get a collection of events based on filter parameters

.DESCRIPTION
Get a collection of events based on filter parameters

.EXAMPLE
PS> Get-EventCollection -Type "my_CustomType2" -DateFrom "-10d"

Get events with type 'my_CustomType' that were created in the last 10 days

.EXAMPLE
PS> Get-EventCollection -Device $Device.id

Get events from a device

.EXAMPLE
PS> Get-DeviceCollection -Name $Device.name | Get-EventCollection

Get events from a device (using pipeline)


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device ID
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # Event type.
        [Parameter()]
        [string]
        $Type,

        # Fragment name from event.
        [Parameter()]
        [string]
        $FragmentType,

        # Start date or date and time of event occurrence.
        [Parameter()]
        [string]
        $DateFrom,

        # End date or date and time of event occurrence.
        [Parameter()]
        [string]
        $DateTo,

        # Return the newest instead of the oldest events. Must be used with dateFrom and dateTo parameters
        [Parameter()]
        [switch]
        $Revert,

        # Maximum number of results
        [Parameter()]
        [AllowNull()]
        [AllowEmptyString()]
        [ValidateRange(1,2000)]
        [int]
        $PageSize,

        # Include total pages statistic
        [Parameter()]
        [switch]
        $WithTotalPages,

        # Get a specific page result
        [Parameter()]
        [int]
        $CurrentPage,

        # Maximum number of pages to retrieve when using -IncludeAll
        [Parameter()]
        [int]
        $TotalPages,

        # Include all results
        [Parameter()]
        [switch]
        $IncludeAll,

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
        $TimeoutSec
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Type")) {
            $Parameters["type"] = $Type
        }
        if ($PSBoundParameters.ContainsKey("FragmentType")) {
            $Parameters["fragmentType"] = $FragmentType
        }
        if ($PSBoundParameters.ContainsKey("DateFrom")) {
            $Parameters["dateFrom"] = $DateFrom
        }
        if ($PSBoundParameters.ContainsKey("DateTo")) {
            $Parameters["dateTo"] = $DateTo
        }
        if ($PSBoundParameters.ContainsKey("Revert")) {
            $Parameters["revert"] = $Revert
        }
        if ($PSBoundParameters.ContainsKey("PageSize")) {
            $Parameters["pageSize"] = $PageSize
        }
        if ($PSBoundParameters.ContainsKey("WithTotalPages") -and $WithTotalPages) {
            $Parameters["withTotalPages"] = $WithTotalPages
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

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference automatic variables
            if (!$WhatIfPreference.IsPresent) {
                $WhatIfPreference = $PSCmdlet.SessionState.PSVariable.get("WhatIfPreference").Value
            }
        
            # Inherit custom parameters
            $Stack = Get-PSCallStack | Select-Object -Skip 1 -First 1
            $InheritVariables = @(@{Source="Force"; Target="Force"}, @{Source="WhatIf"; Target="WhatIfPreference"})
            foreach ($iVariable in $InheritVariables) {
                if (-Not $PSBoundParameters.ContainsKey($iVariable.Source)) {
                    if ($null -ne $Stack -and $Stack.InvocationInfo.BoundParameters.ContainsKey($iVariable.Source)) {
                        Set-Variable -Name $iVariable.Target -Value $Stack.InvocationInfo.BoundParameters[$iVariable.Source] -WhatIf:$false
                    }
                }
            }
        }
    }

    Process {
        $Parameters["device"] = PSc8y\Expand-Id $Device

        if (!$Force -and
            !$WhatIfPreference -and
            !$PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
            )) {
            continue
        }

        Invoke-ClientCommand `
            -Noun "events" `
            -Verb "list" `
            -Parameters $Parameters `
            -Type "application/vnd.com.nsn.cumulocity.eventCollection+json" `
            -ItemType "application/vnd.com.nsn.cumulocity.event+json" `
            -ResultProperty "events" `
            -Raw:$Raw `
            -CurrentPage:$CurrentPage `
            -TotalPages:$TotalPages `
            -IncludeAll:$IncludeAll
    }

    End {}
}
