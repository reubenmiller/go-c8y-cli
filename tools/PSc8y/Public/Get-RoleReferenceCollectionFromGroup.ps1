# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-RoleReferenceCollectionFromGroup {
<#
.SYNOPSIS
Get collection of user role references from a group

.DESCRIPTION
Get collection of user role references from a group

.EXAMPLE
PS> Get-RoleReferenceCollectionFromGroup -Group $Group.id

Get a list of role references for a user group


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Group id (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Group,

        # Tenant
        [Parameter()]
        [object]
        $Tenant,

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
        if ($PSBoundParameters.ContainsKey("Group")) {
            $Parameters["group"] = PSc8y\Expand-Id $Group
        }
        if ($PSBoundParameters.ContainsKey("Tenant")) {
            $Parameters["tenant"] = $Tenant
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

    }

    Process {
        foreach ($item in @("")) {


            Invoke-ClientCommand `
                -Noun "userRoles" `
                -Verb "getRoleReferenceCollectionFromGroup" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.roleReferenceCollection+json" `
                -ItemType "application/vnd.com.nsn.cumulocity.roleReference+json" `
                -ResultProperty "references" `
                -Raw:$Raw
        }
    }

    End {}
}
