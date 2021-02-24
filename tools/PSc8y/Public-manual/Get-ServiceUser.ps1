Function Get-ServiceUser {
    <#
    .SYNOPSIS
    Get service user
    
    .DESCRIPTION
    Get the service user associated to a microservice
    
    .EXAMPLE
    PS> Get-ServiceUser -Id $App.name
    
    Get application service user
    
    #>
        [cmdletbinding(SupportsShouldProcess = $true,
                       PositionalBinding=$true,
                       HelpUri='',
                       ConfirmImpact = 'None')]
        [Alias()]
        [OutputType([object])]
        Param(
            # Microservice id (required)
            [Parameter(Mandatory = $true,
                       ValueFromPipeline=$true,
                       ValueFromPipelineByPropertyName=$true)]
            [object[]]
            $Id,
    
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
            $Timeout
        )
    
        Begin {
            $Parameters = @{}
            if ($PSBoundParameters.ContainsKey("OutputFile")) {
                $Parameters["outputFile"] = $OutputFile
            }
            if ($PSBoundParameters.ContainsKey("NoProxy")) {
                $Parameters["noProxy"] = $NoProxy
            }
            if ($PSBoundParameters.ContainsKey("Session")) {
                $Parameters["session"] = $Session
            }
            if ($PSBoundParameters.ContainsKey("Timeout")) {
                $Parameters["timeout"] = $Timeout
            }
        }
    
        Process {
            foreach ($item in (PSc8y\Expand-Microservice $Id)) {
                if ($item) {
                    $Parameters["id"] = if ($item.id) { $item.id } else { $item }
                }
    
                Invoke-ClientCommand `
                    -Noun "microservices" `
                    -Verb "getServiceUser" `
                    -Parameters $Parameters `
                    -Type "application/vnd.com.nsn.cumulocity.applicationUserCollection+json" `
                    -ItemType "application/vnd.com.nsn.cumulocity.bootstrapuser+json" `
                    -ResultProperty "users" `
                    -Raw:$Raw
            }
        }
    
        End {}
    }
    