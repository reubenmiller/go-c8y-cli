# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-ChildAssetCollection {
<#
.SYNOPSIS
Get a collection of managedObjects child references

.DESCRIPTION
Get a collection of managedObjects child references

.EXAMPLE
PS> Get-ChildAssetCollection -Group $Group.id
Get a list of the child assets of an existing device

.EXAMPLE
PS> Get-ChildAssetCollection -Group $Group.id
Get a list of the child assets of an existing group


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device.
        [Parameter()]
        [object[]]
        $Device,

        # Group.
        [Parameter()]
        [object[]]
        $Group,

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
        if ($PSBoundParameters.ContainsKey("Device")) {
            $Parameters["device"] = $Device
        }
        if ($PSBoundParameters.ContainsKey("Group")) {
            $Parameters["group"] = $Group
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
                -Noun "inventoryReferences" `
                -Verb "listChildAssets" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.managedObjectReferenceCollection+json" `
                -ItemType "application/vnd.com.nsn.cumulocity.managedObject+json" `
                -ResultProperty "references.managedObject" `
                -Raw:$Raw
        }
    }

    End {}
}
