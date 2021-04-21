# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-ExternalId {
<#
.SYNOPSIS
Delete external id

.DESCRIPTION
Delete an existing external id. This does not delete the device managed object

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/identity_delete

.EXAMPLE
PS> Remove-ExternalId -Type "my_SerialNumber" -Name "myserialnumber2"

Delete external identity


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # External identity type
        [Parameter()]
        [string]
        $Type,

        # External identity id/name (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Name
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "identity delete"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Name `
            | Group-ClientRequests `
            | c8y identity delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Name `
            | Group-ClientRequests `
            | c8y identity delete $c8yargs
        }
        
    }

    End {}
}
