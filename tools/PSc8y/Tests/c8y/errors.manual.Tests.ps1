. $PSScriptRoot/../imports.ps1

Describe -Name "c8y errors" {
    BeforeAll {
        $ids = New-Object System.Collections.ArrayList
    }

    Context "server errors" {
        It "returns an empty response for server errors" {
            $output = @("0") | c8y events get
            $LASTEXITCODE | Should -Be 4
            $output | Should -BeNullOrEmpty
        }
    }

    Context "command errors" {
        It "returns an empty response for command errors" {
            $output = c8y events get --iiiiid 0
            $LASTEXITCODE | Should -Be 100
            $output | Should -BeNullOrEmpty
        }

        It "returns writes errors to stderr" {
            $output = c8y events get --iiiiid 0 2>&1
            $LASTEXITCODE | Should -Be 100
            $output | Should -Not -BeNullOrEmpty
            $output | Should -Match "ERROR\s+commandError: unknown flag: --iiiiid"
        }

        It "returns errors as json" {
            $output = c8y events get --withError --iiiiid 0
            $LASTEXITCODE | Should -Be 100

            $output | Should -Not -BeNullOrEmpty
            $details = ConvertFrom-Json $output -Depth 100
            $details | Should -MatchObject @{errorType="commandError"; message = "unknown flag: --iiiiid"}
        }

        It "writes errors to stdout when using withError" {
            # Note: --withError must be included before the invalid flag
            $output = c8y events get --withError --iiiiid 0
            $LASTEXITCODE | Should -Be 100

            $output | Should -Not -BeNullOrEmpty
            $details = ConvertFrom-Json $output -Depth 100
            $details | Should -MatchObject @{errorType="commandError"; message = "unknown flag: --iiiiid"}
        }

        It "silences specific status codes as the user knows that error might not occur and is ok with it" {
            $output = c8y events get --id 0 --silentStatusCodes 404 2>&1
            $LASTEXITCODE | Should -Be 4
            $output | Should -BeNullOrEmpty -Because "The user does not want to return any errors"
        }

        It "silences specific status codes also when reference by name is being used" {
            $output = c8y devices get --id myNonExistantDevice --silentStatusCodes 404 2>&1
            $LASTEXITCODE | Should -Be 4
            $output | Should -BeNullOrEmpty -Because "The user does not want to return any errors"
        }

        It "returns a timeout error" {
            $retries = 3
            $exitCode = $null
            do {
                $output = c8y devices list --timeout "1ms" --withError
                $exitCode = $LASTEXITCODE

                if ($exitCode -eq 106) {
                    break
                }
                $retries -= 1
                Start-Sleep -Seconds 1
                
            } while ($retries -gt 0)

            $exitCode | Should -Be 106

            $output | Should -Not -BeNullOrEmpty
            $details = ConvertFrom-Json $output -Depth 100
            $details | Should -MatchObject @{errorType="commandError"; exitCode = 106; message = "command timed out"}
        }        
    }

    AfterAll {
        $ids | Remove-ManagedObject
    }
}
