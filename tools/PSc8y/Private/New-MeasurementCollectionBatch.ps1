function New-MeasurementCollectionBatch {
    <#
    .SYNOPSIS
    Create a measurement for a collection of devices using client side batching
    
    .DESCRIPTION
    Create a measurement for a collection of devices using client side batching
    
    .EXAMPLE
    PS> New-MeasurementCollectionBatch -InputFile "mylist.csv" -Data @{ custom_data = @{ value = 1 } } -WhatIf
    
    Using WhatIf, check what requests will be sent to create a measurement on each of the listed device ids in a given file list (no data is transmitted to the server)
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
                    -Verb "createMeasurements" `
                    -Parameters $Parameters `
                    -Type "application/vnd.com.nsn.cumulocity.measurement+json" `
                    -ItemType "" `
                    -ResultProperty "" `
                    -Raw:$Raw
            }
        }
    
        end {
    
        }
    }