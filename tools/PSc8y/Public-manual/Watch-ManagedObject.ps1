﻿Function Watch-ManagedObject {
<#
.SYNOPSIS
Watch realtime managedObjects

.DESCRIPTION
Watch realtime managedObjects

.EXAMPLE
PS> Watch-ManagedObject -Device 12345
Watch all managedObjects for a device

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

        # Start date or date and time of managedObject occurrence. (required)
        [Parameter()]
        [int]
        $DurationSec,

        # End date or date and time of managedObject occurrence.
        [Parameter()]
        [string]
        $Count,

        # Outputfile
        [Parameter()]
        [string]
        $OutputFile,

        # NoProxy
        [Parameter()]
        [switch]
        $NoProxy,

        # Session path
        [Parameter()]
        [string]
        $Session
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("DurationSec")) {
            $Parameters["duration"] = $DurationSec
        }
        if ($PSBoundParameters.ContainsKey("Count")) {
            $Parameters["count"] = $Count
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

    }

    Process {
        $id = PSc8y\Expand-Id $Device
        if ($id) {
            $Parameters["device"] = PSc8y\Expand-Id $Device
        }

        if (!$Force -and
            !$WhatIfPreference -and
            !$PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
            )) {
            continue
        }

        $c8yargs = New-Object System.Collections.ArrayList
        $null = $c8yargs.AddRange(@("inventory", "subscribe"))
        $Parameters.Keys | ForEach-Object {
            $null = $c8yargs.AddRange(@("$_", $Parameters[$_]))
        }

        Invoke-BinaryProcess (Get-ClientBinary) -RedirectOutput -ArgumentList $c8yargs |
            Add-PowershellType "application/json"
    }

    End {}
}
