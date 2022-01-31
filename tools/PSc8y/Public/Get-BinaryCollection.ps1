# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-BinaryCollection {
<#
.SYNOPSIS
Get binary collection

.DESCRIPTION
Get a collection of inventory binaries. The results include the meta information about binary and not the binary itself.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/binaries_list

.EXAMPLE
PS> Get-BinaryCollection -PageSize 100

Get a list of binaries


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # List of ids.
        [Parameter()]
        [string[]]
        $Ids,

        # ManagedObject type.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Type,

        # ManagedObject fragment type.
        [Parameter()]
        [string]
        $FragmentType,

        # List of managed objects that are owned by the given username.
        [Parameter()]
        [string]
        $Owner,

        # managed objects containing a text value starting with the given text (placeholder {text}). Text value is any alphanumeric string starting with a latin letter (A-Z or a-z).
        [Parameter()]
        [string]
        $Text,

        # Search for a specific child addition and list all the groups to which it belongs.
        [Parameter()]
        [string]
        $ChildAdditionId,

        # Search for a specific child asset and list all the groups to which it belongs.
        [Parameter()]
        [string]
        $ChildAssetId,

        # Search for a specific child device and list all the groups to which it belongs.
        [Parameter()]
        [object[]]
        $ChildDeviceId
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get", "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "binaries list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Type `
            | Group-ClientRequests `
            | c8y binaries list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Type `
            | Group-ClientRequests `
            | c8y binaries list $c8yargs
        }
        
    }

    End {}
}
