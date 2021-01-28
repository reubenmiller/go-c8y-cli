function New-ManagedObjectCollectionBatch {
<#
.SYNOPSIS
Create a collection of new inventory managed objects using client side batching

.DESCRIPTION
Create a collection of new inventory managed objects using client side batching

.EXAMPLE
PS> New-ManagedObjectCollectionBatch -Count 5 -Data @{ custom_data = @{ value = 1 } } -WhatIf

Using WhatIf, check what requests will be sent to create 5 managed objects (no data is transmitted to the server)
#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    param (
        # Additional properties of the inventory.
        [Parameter()]
        [object]
        $Data,

        # Template (jsonnet) file to use to create the request body.
        [Parameter()]
        [string]
        $Template,

        # Variables to be used when evaluating the Template. Accepts a file path, json or json shorthand, i.e. "name=peter"
        [Parameter()]
        [string]
        $TemplateVars,

        # Number of workers
        [Parameter()]
        [int]
        $Workers,

        # Number of objects to create
        [Parameter()]
        [int]
        $Count,

        # Start index. Range will go from startIndex to startIndex + Count. Default to 1
        [Parameter()]
        [int]
        $StartIndex,

        # Delay in milliseconds after each request before the worker is released back to the pool
        [Parameter()]
        [int]
        $Delay,

        # Abort batch if the error count reaches or exceeds this amount. Defaults to 10
        [Parameter()]
        [int]
        $AbortOnErrors,

        # Show the full (raw) response from Cumulocity including pagination information
        [Parameter()]
        [switch]
        $Raw,

        # Cumulocity processing mode
        [Parameter()]
        [AllowNull()]
        [AllowEmptyString()]
        [ValidateSet("PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP", "")]
        [string]
        $ProcessingMode,

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
    
    begin {
        $Parameters = @{}
        if ($PSBoundParameters.ContainsKey("Workers")) {
            $Parameters["workers"] = $Workers
        }
        if ($PSBoundParameters.ContainsKey("Count")) {
            $Parameters["count"] = $Count
        }
        if ($PSBoundParameters.ContainsKey("StartIndex")) {
            $Parameters["startIndex"] = $StartIndex
        }
        if ($PSBoundParameters.ContainsKey("Delay")) {
            $Parameters["delay"] = $Delay
        }
        if ($PSBoundParameters.ContainsKey("AbortOnErrors")) {
            $Parameters["abortOnErrors"] = $AbortOnErrors
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
    
    process {
        foreach ($item in @("")) {

            if (!$Force -and
                !$WhatIfPreference -and
                !$PSCmdlet.ShouldProcess(
                    (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                    (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $item)
                )) {
                continue
            }

            Invoke-ClientCommand `
                -Noun "batch" `
                -Verb "createManagedObjects" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.inventory+json" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }
    
    end {
        
    }
}