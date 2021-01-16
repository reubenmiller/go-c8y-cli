# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-RetentionRule {
<#
.SYNOPSIS
Update retention rule

.DESCRIPTION
Update an existing retentule rule, i.e. change maximum number of days or the data type.


.EXAMPLE
PS> Update-RetentionRule -Id $RetentionRule.id -DataType MEASUREMENT -FragmentType "custom_FragmentType"

Update a retention rule

.EXAMPLE
PS> Get-RetentionRule -Id $RetentionRule.id | Update-RetentionRule -DataType MEASUREMENT -FragmentType "custom_FragmentType"

Update a retention rule (using pipeline)


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Retention rule id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [string]
        $Id,

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

        # Maximum age of document in days.
        [Parameter()]
        [long]
        $MaximumAge,

        # Whether the rule is editable. Can be updated only by management tenant.
        [Parameter()]
        [switch]
        $Editable,

        # Data
        [Parameter()]
        [object]
        $Data,

        # Cumulocity processing mode
        [Parameter()]
        [AllowNull()]
        [AllowEmptyString()]
        [ValidateSet("PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP", "")]
        [string]
        $ProcessingMode,

        # Template (jsonnet) file to use to create the request body.
        [Parameter()]
        [string]
        $Template,

        # Variables to be used when evaluating the Template. Accepts a file path, json or json shorthand, i.e. "name=peter"
        [Parameter()]
        [string]
        $TemplateVars,

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
        $TimeoutSec,

        # Don't prompt for confirmation
        [Parameter()]
        [switch]
        $Force
    )

    Begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("DataType")) {
            $Parameters["dataType"] = $DataType
        }
        if ($PSBoundParameters.ContainsKey("FragmentType")) {
            $Parameters["fragmentType"] = $FragmentType
        }
        if ($PSBoundParameters.ContainsKey("Type")) {
            $Parameters["type"] = $Type
        }
        if ($PSBoundParameters.ContainsKey("Source")) {
            $Parameters["source"] = $Source
        }
        if ($PSBoundParameters.ContainsKey("MaximumAge")) {
            $Parameters["maximumAge"] = $MaximumAge
        }
        if ($PSBoundParameters.ContainsKey("Editable")) {
            $Parameters["editable"] = $Editable
        }
        if ($PSBoundParameters.ContainsKey("Data")) {
            $Parameters["data"] = ConvertTo-JsonArgument $Data
        }
        if ($PSBoundParameters.ContainsKey("ProcessingMode")) {
            $Parameters["processingMode"] = $ProcessingMode
        }
        if ($PSBoundParameters.ContainsKey("Template") -and $Template) {
            $Parameters["template"] = $Template
        }
        if ($PSBoundParameters.ContainsKey("TemplateVars") -and $TemplateVars) {
            $Parameters["templateVars"] = $TemplateVars
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

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }
    }

    Process {
        foreach ($item in (PSc8y\Expand-Id $Id)) {
            if ($item) {
                $Parameters["id"] = if ($item.id) { $item.id } else { $item }
            }

            if (!$Force -and
                !$WhatIfPreference -and
                !$PSCmdlet.ShouldProcess(
                    (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                    (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
                )) {
                continue
            }

            Invoke-ClientCommand `
                -Noun "retentionRules" `
                -Verb "update" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.retentionRule+json" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }

    End {}
}
