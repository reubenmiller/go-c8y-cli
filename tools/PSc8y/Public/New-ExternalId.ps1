# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-ExternalId {
<#
.SYNOPSIS
Create external id

.DESCRIPTION
Create a new external id for an existing managed object

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/identity_create

.EXAMPLE
PS> New-ExternalId -Device $Device.id -Type "$my_SerialNumber" -Name "myserialnumber"

Create external identity

.EXAMPLE
PS> Get-Device $Device.id | New-ExternalId -Type "$my_SerialNumber" -Template "{ externalId: input.value.name }"

Create external identity (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # The ManagedObject linked to the external ID. (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Device,

        # The type of the external identifier as string, e.g. 'com_cumulocity_model_idtype_SerialNumber'.
        [Parameter()]
        [string]
        $Type,

        # The identifier used in the external system that Cumulocity interfaces with.
        [Parameter()]
        [string]
        $Name
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "identity create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.externalId+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Device `
            | Group-ClientRequests `
            | c8y identity create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Device `
            | Group-ClientRequests `
            | c8y identity create $c8yargs
        }
        
    }

    End {}
}
