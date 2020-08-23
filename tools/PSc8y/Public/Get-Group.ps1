# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-Group {
<#
.SYNOPSIS
Create a new group by id

.EXAMPLE
PS> Get-Group -Id $Group.id
Get a user group


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Group id
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # Tenant
        [Parameter()]
        [object]
        $Tenant,

        # Include raw response including pagination information
        [Parameter()]
        [switch]
        $Raw,

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
        $Session,

        # TimeoutSec timeout in seconds before a request will be aborted
        [Parameter()]
        [double]
        $TimeoutSec
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Tenant")) {
            $Parameters["tenant"] = $Tenant
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
        $Parameters["id"] = PSc8y\Expand-Id $Id

        if (!$Force -and
            !$WhatIfPreference -and
            !$PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
            )) {
            continue
        }

        Invoke-ClientCommand `
            -Noun "userGroups" `
            -Verb "get" `
            -Parameters $Parameters `
            -Type "application/vnd.com.nsn.cumulocity.group+json" `
            -ItemType "" `
            -ResultProperty "" `
            -Raw:$Raw `
            -IncludeAll:$IncludeAll
    }

    End {}
}
