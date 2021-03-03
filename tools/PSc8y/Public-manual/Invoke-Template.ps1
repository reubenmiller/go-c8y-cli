Function Invoke-Template {
    <#
    .SYNOPSIS
    Execute a jsonnet data template
    
    .DESCRIPTION
    Execute a jsonnet data template and show the output of the template. Useful when developing new templates
    
    .LINK
    c8y template execute

    .EXAMPLE
    PS> Invoke-Template -Template ./template.jsonnet
    
    Execute a jsonnet template

    .EXAMPLE
    PS> Invoke-Template -Template ./template.jsonnet -TemplateVars "name=input"
    
    Execute a jsonnet template

    .EXAMPLE
    PS> Invoke-Template -Template ./template.jsonnet -TemplateVars "name=input,type=mytype"
    
    Execute a jsonnet template which has multiple template variables (using a comma separated string)

    .OUTPUTS
    String
    
    #>
    [cmdletbinding(SupportsShouldProcess = $false,
        PositionalBinding = $true,
        HelpUri = '',
        ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(    
        # Template (jsonnet) file to use to create the request body.
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [string]
        $Template,
    
        # Variables to be used when evaluating the Template. Accepts a file path, json or json shorthand, i.e. "name=peter"
        [Parameter()]
        [string]
        $TemplateVars,

        # Template input data
        [Parameter(
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true
        )]
        [object[]]
        $Data,
            
        # Output compressed/minified json
        [switch] $Compress
    )
    
    Begin {
        $c8yArgs = New-Object System.Collections.ArrayList

        if ($PSBoundParameters.ContainsKey("Template") -and $Template) {
            $null = $c8yArgs.AddRange(@("--template", $Template))
        }
        if ($PSBoundParameters.ContainsKey("TemplateVars") -and $TemplateVars) {
            $null = $c8yArgs.AddRange(@("--templateVars", $TemplateVars))
        }

        if ($Compress) {
            $null = $c8yArgs.Add("--compact=true")
        }
    }
    
    Process {
        $InputData = @($null)

        if ($null -ne $Data) {
            $InputData = $Data
        }

        foreach ($iData in $InputData) {
            $ic8yArgs = $c8yArgs.Clone()

            if ($iData) {
                $null = $ic8yArgs.AddRange(@("--data", (ConvertTo-JsonArgument $iData)))
            }

            c8y template execute $ic8yArgs
        }
    }
    
    End {}
}
