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

        It "returns writes errors as json to stderr" {
            $output = c8y events get --iiiiid 0 2>&1
            $LASTEXITCODE | Should -Be 100

            $output | Should -Not -BeNullOrEmpty
            $details = ConvertFrom-Json $output -Depth 100
            $details | Should -MatchObject @{error="commandError"; message = "unknown flag: --iiiiid"}
        }

        It "returns writes errors to stdout when using withError" {
            # Note: --withError must be included before the invalid flag
            $output = c8y events get --withError --iiiiid 0
            $LASTEXITCODE | Should -Be 100

            $output | Should -Not -BeNullOrEmpty
            $details = ConvertFrom-Json $output -Depth 100
            $details | Should -MatchObject @{error="commandError"; message = "unknown flag: --iiiiid"}
        }

        It "silencies specific status codes as the user knows that error might not occur and is ok with it" {
            $output = c8y events get --id 0 --silentStatusCodes 404 2>&1
            $LASTEXITCODE | Should -Be 4
            $output | Should -BeNullOrEmpty -Because "The user does not want to return any errors"
        }

        It "returns a timeout error" {
            $output = c8y devices list --timeout 0.001 --withError
            $LASTEXITCODE | Should -Be 106

            $output | Should -Not -BeNullOrEmpty
            $details = ConvertFrom-Json $output -Depth 100
            $details | Should -MatchObject @{error="commandError"; message = "command timed out"}
        }        
    }

    AfterAll {
        $ids | Remove-ManagedObject
    }
}
