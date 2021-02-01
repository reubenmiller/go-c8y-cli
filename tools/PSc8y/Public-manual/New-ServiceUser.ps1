Function New-ServiceUser {
    <#
    .SYNOPSIS
    New service user (via a dummy microservice user)

    .DESCRIPTION
    Create a new microservice application used to provide a service user used for external automation tasks

    .EXAMPLE
    PS> New-ServiceUser -Name "automation01" -RequiredRoles @("ROLE_INVENTORY_READ") -Tenants t123456

    Create a new microservice called automation01 which has permissions to read the inventory, and subscribe the application to tenant t123456

    .LINK
    Get-ServiceUser
    #>
        [cmdletbinding(SupportsShouldProcess = $true,
                       PositionalBinding=$true,
                       HelpUri='',
                       ConfirmImpact = 'High')]
        [Alias()]
        [OutputType([object])]
        Param(
            # Name of the microservice. An id is also accepted however the name have been previously uploaded.
            [Parameter(
                Mandatory = $true,
                Position = 0
            )]
            [string]
            $Name,

            # Roles which should be assigned to the service user, i.e. ROLE_INVENTORY_READ
            [Parameter(Mandatory = $false)]
            [string[]]
            $Roles,

            # Tenants IDs where the application should be subscribed. Useful when using in a multi tenant scenario where the
            # application is created in the management tenant, and a service user can be created via subscribing to the application on each
            # sub tenant
            [Parameter(Mandatory = $false)]
            [string[]]
            $Tenants,

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
            $TimeoutSec,

            # Don't prompt for confirmation
            [Parameter()]
            [switch]
            $Force
        )

        Begin {
            $Parameters = @{}
            if ($PSBoundParameters.ContainsKey("Name")) {
                $Parameters["name"] = $Name
            }
            if ($PSBoundParameters.ContainsKey("Roles")) {
                $Parameters["roles"] = $Roles -join ","
            }
            if ($PSBoundParameters.ContainsKey("Tenants")) {
                $Parameters["tenants"] = $Tenants -join ","
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

            foreach ($item in @($Name)) {

                $Parameters["name"] = $item

                if (!$Force -and
                    !$WhatIfPreference -and
                    !$PSCmdlet.ShouldProcess(
                        (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                        (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
                    )) {
                    continue
                }

                Invoke-ClientCommand `
                    -Noun "microservices" `
                    -Verb "createServiceUser" `
                    -Parameters $Parameters `
                    -Type "application/vnd.com.nsn.cumulocity.application+json" `
                    -ItemType "" `
                    -ResultProperty "" `
                    -Raw:$Raw
            }
        }

        End {}
    }
