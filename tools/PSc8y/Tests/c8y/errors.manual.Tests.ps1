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

        It "prints errors on stderr and exit code when a single error" {
            $id = c8y inventory create --output csv --select id
            [void] $ids.Add($id)
            $stderr = $( $output = @($id, "0") | c8y inventory get ) 2>&1
            $LASTEXITCODE | Should -Be 4 -Because "single 404 error"
            $mos = $output | ConvertFrom-Json
            $mos[0].id | Should -BeExactly $id
            $stderr | Should -HaveCount 1
            $stderr | Should -BeLike "*404*"
        }

        It "silences errors for specific status codes on stderr" {
            $id = c8y inventory create --output csv --select id
            [void] $ids.Add($id)
            $stderr = $( $output = @($id, "0") | c8y inventory get --silentStatusCodes 404 ) 2>&1
            $LASTEXITCODE | Should -Be 4 -Because "all 404 errors have been silenced"
            $mos = $output | ConvertFrom-Json
            $mos[0].id | Should -BeExactly $id
            $stderr | Should -HaveCount 0 -Because "Silent status codes do not get logged"
            $stderr | Should -Not -BeLike "*404*"
        }

        It "silences errors for specific status codes on stderr" {
            $id = c8y inventory create --output csv --select id
            [void] $ids.Add($id)
            $stderr = $( $output = @($id, "0") | c8y inventory get --silentStatusCodes 404 --silentExit ) 2>&1
            $LASTEXITCODE | Should -Be 0 -Because "silent status codes should not affect exit code"
            $mos = $output | ConvertFrom-Json
            $mos[0].id | Should -BeExactly $id
            $stderr | Should -HaveCount 0 -Because "Silent status codes do not get logged"
        }

        It "silences errors for specific status codes on stderr" {
            $id = c8y inventory create --output csv --select id
            [void] $ids.Add($id)
            $stderr = $( $output = @($id, "0", "0") | c8y inventory get ) 2>&1
            $LASTEXITCODE | Should -Be 104 -Because "completed with errors"
            $mos = $output | ConvertFrom-Json
            $mo | Should -HaveCount 1
            $mos[0].id | Should -BeExactly $id
            $stderr | Should -HaveCount 3 -Because "2 x 404 errors + 1 summary error"
            $stderr -like "*404*" | Should -HaveCount 2
            $stderr -like "*completed with 2 errors*" | Should -HaveCount 1
        }

        It "supports ternary operations on the exit code" {
            $name = New-RandomString

            # check if device exists by name if not create it
            $output = c8y devices get --id $name --silentStatusCodes 404 || c8y devices create --name $name

            $mo = $output | ConvertFrom-Json
            $mo[0].name | Should -BeExactly $name

            # run the command again, this time the name should exist so it should not be created
            $output = c8y devices get --id $name --silentStatusCodes 404 || c8y devices create --name $name
            $mo2 = $output | ConvertFrom-Json
            $mo2.id | Should -BeExactly $mo.id -Because "id should already exist"
        }

        It "deletes multiple ids but dont care if they dont exist" {
            $id = c8y inventory create --select id --output csv

            # check if device exists by name if not create it
            $output = @("0", $id) | c8y inventory delete --silentStatusCodes 404 2>&1
            $LASTEXITCODE | Should -Be 4
            $output | Should -HaveCount 0

            $output = @("0", "0") | c8y inventory delete --silentStatusCodes 404 --silentExit 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -HaveCount 0

            $output = @("0", "0") | c8y inventory delete 2>&1
            $LASTEXITCODE | Should -Be 104
            $output | Should -HaveCount 3 -Because "2 x errors + 1 summary error"
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
        
        It "return an error if required body properties are missing" {
            $stderr = $( $output = c8y events create --device 12345 --type c8y_TestAlarm --force ) 2>&1
            $LASTEXITCODE | Should -Be 101
            $output | Should -BeNullOrEmpty
            $stderr | Should -HaveCount 1
            $stderr[-1] -match "Body is missing required properties: text" | Should -HaveCount 1
        }

        It "handles multiple errors in pipeline" {
            $stderr = $( $output = "0", "0" | c8y events create --type "c8y_TestAlarm" --force ) 2>&1
            $LASTEXITCODE | Should -Be 104
            $output | Should -BeNullOrEmpty
            $stderr | Should -HaveCount 3
            $stderr[0] -match "Body is missing required properties: text" | Should -HaveCount 1
            $stderr[1] -match "Body is missing required properties: text" | Should -HaveCount 1
            $stderr[2] -match "jobs completed with 2 errors" | Should -HaveCount 1
        }

        It "handles multiple errors in pipeline and when mode is not set properly" {
            $stderr = $( $output = "0", "0" | c8y events create --type "c8y_TestAlarm" --force ) 2>&1
            $LASTEXITCODE | Should -Be 104
            $output | Should -BeNullOrEmpty
            $stderr | Should -HaveCount 3
            $stderr[0] -match "Body is missing required properties: text" | Should -HaveCount 1
            $stderr[1] -match "Body is missing required properties: text" | Should -HaveCount 1
            $stderr[2] -match "jobs completed with 2 errors" | Should -HaveCount 1
        }
    }

    AfterAll {
        $ids | Remove-ManagedObject
    }
}
