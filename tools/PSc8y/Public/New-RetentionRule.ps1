# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-RetentionRule {
<#
.SYNOPSIS
Create retention rule

.DESCRIPTION
Create a new retention rule to managed when data is deleted in the tenant


.LINK
c8y retentionrules create

.EXAMPLE
PS> New-RetentionRule -DataType ALARM -MaximumAge 180

Create a retention rule to delete all alarms after 180 days


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # RetentionRule will be applied to this type of documents, possible values [ALARM, AUDIT, EVENT, MEASUREMENT, OPERATION, *]. (required)
        [Parameter(Mandatory = $true)]
        [ValidateSet('ALARM','AUDIT','EVENT','MEASUREMENT','OPERATION','*')]
        [string]
        $DataType,

        # RetentionRule will be applied to documents with fragmentType.
        [Parameter()]
        [string]
        $FragmentType,

        # RetentionRule will be applied to documents with type.
        [Parameter()]
        [string]
        $Type,

        # RetentionRule will be applied to documents with source.
        [Parameter()]
        [string]
        $Source,

        # Maximum age of document in days. (required)
        [Parameter(Mandatory = $true)]
        [long]
        $MaximumAge,

        # Whether the rule is editable. Can be updated only by management tenant.
        [Parameter()]
        [switch]
        $Editable
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "retentionrules create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.retentionRule+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y retentionrules create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y retentionrules create $c8yargs
        }
    }

    End {}
}
