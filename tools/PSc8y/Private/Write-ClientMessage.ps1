Function Write-ClientMessage {
<# 

.OUTPUTS
[[]string] Remaining non-log messages
#>
    [cmdletbinding(
        SupportsShouldProcess = $true,
        ConfirmImpact = "None"
    )]
    Param(
        # Input message
        [Parameter(
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true,
            Position = 0
        )]
        [AllowEmptyString()]
        [string[]] $InputObject,

        [switch] $PassThru
    )
    Process {
        $date = Get-Date -Format "yyyy-MM-dd"
        $detectedPanic = $false
        $panic = $null

        $CommandInfo = @{}

        foreach ($line in $InputObject) {

            # Extra data about the request
            if (!$WhatIfPreference) {
                switch -Regex ($line) {
                    "Response time: " {
                        $value = $line -split "Response time: " | Select-Object -Last 1
                        $CommandInfo["responseTime"] = $value
                    }
                    "Status code:" {
                        $value = $line -split "Status code: " | Select-Object -Last 1
                        $CommandInfo["statusCode"] = $value
                    }
                    "Response header:" {
                        $value = $line -split "Response header: " | Select-Object -Last 1
                        $CommandInfo["responseHeader"] = $value
                    }
                    "Response Length:" {
                        $value = $line -split "Response Length: " | Select-Object -Last 1
                        $CommandInfo["responseLength"] = $value
                    }
                }
            }
            
            if ($line -eq "System.Management.Automation.RemoteException" -or $line -eq "System.Management.Automation.CmdletInvocationException") {
                $line = ""
                if ($WhatIfPreference) {
                    Write-InformationColored -MessageData $line -ForegroundColor Green -ShowHost
                } else {
                    if ($PassThru) {
                        Write-Output $line
                    }
                }
                continue
            }

            if ($line.StartsWith("panic")) {
                $detectedPanic = $true
                $panic = New-Object System.Collections.ArrayList
            }

            if (-Not $line.StartsWith($date)) {
                if ($WhatIfPreference -and (-Not $line.Contains('"error":"commandError"'))) {
                    Write-InformationColored -MessageData $line -ForegroundColor Green
                } else {
                    if ($detectedPanic) {
                        $null = $panic.Add($line)
                    } else {
                        if ($PassThru) {
                            Write-Output $line
                        }
                    }
                }
                continue
            }

            switch -Regex -CaseSensitive ($line) {
                "What If:" {
                    $line = $line -replace ".*(What if:)", '$1'
                    Write-InformationColored -MessageData "$line" -ForegroundColor Green -ShowHost
                }

                "INFO" {
                    Write-Verbose "$line"
                }

                "DEBUG" {
                    Write-Verbose "$line"
                }

                "WARN" {
                    Write-Warning "$line"
                }

                "ERROR" {
                    Write-Warning "$line"
                }

                "RemoteException" {
                    # write empty string
                    Write-InformationColored -MessageData "" -ForegroundColor Green -ShowHost
                }

                Default {
                    # Remaining WhatIf content
                    Write-InformationColored -MessageData "$line" -ForegroundColor Green -ShowHost
                }
            }
        }

        if ($null -ne $panic) {
            Write-Error ($panic -join "`n")
        }

        if ($CommandInfo.Keys.Count -gt 0) {
            Write-Information -MessageData ([pscustomobject]$CommandInfo) -Tags "response"
        }
        
    }
}