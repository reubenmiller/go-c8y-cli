# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-BulkOperation {
<#
.SYNOPSIS
Create a new bulk operation

.DESCRIPTION
Create a new bulk operation

.EXAMPLE
PS> New-BulkOperation -Group $group.id -StartDate "10s" -CreationRampSec 15 -Operation @{ c8y_Restart = @{} }

Create bulk operation for a group

.EXAMPLE
PS> Get-DeviceGroup $group.id | New-BulkOperation -StartDate "10s" -CreationRampSec 15 -Operation @{ c8y_Restart = @{} }

Create bulk operation for a group (using pipeline)


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Identifies the target group on which this operation should be performed. (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Group,

        # Time when operations should be created.
        [Parameter()]
        [string]
        $StartDate,

        # Delay between every operation creation. (required)
        [Parameter(Mandatory = $true)]
        [long]
        $CreationRampSec,

        # Operation prototype to send to each device in the group (required)
        [Parameter(Mandatory = $true)]
        [object]
        $Operation,

        # Additional properties describing the bulk operation which will be performed on the device group.
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
        if ($PSBoundParameters.ContainsKey("StartDate")) {
            $Parameters["startDate"] = $StartDate
        }
        if ($PSBoundParameters.ContainsKey("CreationRampSec")) {
            $Parameters["creationRampSec"] = $CreationRampSec
        }
        if ($PSBoundParameters.ContainsKey("Operation")) {
            $Parameters["operation"] = ConvertTo-JsonArgument $Operation
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
        foreach ($item in (PSc8y\Expand-Id $Group)) {
            if ($item) {
                $Parameters["group"] = if ($item.id) { $item.id } else { $item }
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
                -Noun "bulkOperations" `
                -Verb "create" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.bulkoperation+json" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }

    End {}
}
