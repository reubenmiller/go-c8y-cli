﻿# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-RetentionRule {
<#
.SYNOPSIS
Update retention rule

.DESCRIPTION
Update an existing retention rule, i.e. change maximum number of days or the data type.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/retentionrules_update

.EXAMPLE
PS> Update-RetentionRule -Id $RetentionRule.id -DataType MEASUREMENT -FragmentType "custom_FragmentType"

Update a retention rule

.EXAMPLE
PS> Get-RetentionRule -Id $RetentionRule.id | Update-RetentionRule -DataType MEASUREMENT -FragmentType "custom_FragmentType"

Update a retention rule (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Retention rule id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # RetentionRule will be applied to this type of documents, possible values [ALARM, AUDIT, EVENT, MEASUREMENT, OPERATION, *].
        [Parameter()]
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

        # Maximum age of document in days.
        [Parameter()]
        [long]
        $MaximumAge,

        # Whether the rule is editable. Can be updated only by management tenant.
        [Parameter()]
        [switch]
        $Editable
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "retentionrules update"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.retentionRule+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y retentionrules update $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y retentionrules update $c8yargs
        }
        
    }

    End {}
}
