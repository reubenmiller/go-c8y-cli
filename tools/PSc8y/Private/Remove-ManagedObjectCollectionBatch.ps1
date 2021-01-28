function Remove-ManagedObjectCollectionBatch {
<#
.SYNOPSIS
Remove a collection of new inventory managed objects using client side batching

.DESCRIPTION
Remove a collection of new inventory managed objects using client side batching

.EXAMPLE
PS> Remove-ManagedObjectCollectionBatch -InputFile "mylist.csv" -Data @{ custom_data = @{ value = 1 } } -WhatIf

Using WhatIf, check what requests will be sent to update all managed object ids in the given file list (no data is transmitted to the server)
#>
    [cmdletbinding(SupportsShouldProcess = $true,
        PositionalBinding = $true,
        HelpUri = '',
        ConfirmImpact = 'High')]
    param (
        # Input file which contains the list of managed object ids to be updated.
        [Parameter(
            Mandatory = $true,
            Position = 0,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true)]
        [Alias("FullName")]
        [object[]]
        $InputFile,
    
        # Number of workers
        [Parameter()]
        [int]
        $Workers,
    
        # Number of objects to create
        # [Parameter()]
        # [int]
        # $Count,
    
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
        # if ($PSBoundParameters.ContainsKey("Count")) {
        #     $Parameters["count"] = $Count
        # }
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
        foreach ($item in $InputFile) {
            $Parameters["inputFile"] = $item
    
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
                -Verb "deleteManagedObjects" `
                -Parameters $Parameters `
                -Type "" `
                -ItemType "" `
                -ResultProperty "" `
                -Raw:$Raw
        }
    }
    
    end {
    
    }
}