. $PSScriptRoot/imports.ps1

Describe -Name "c8y pipes" {
    BeforeAll {
        Set-Alias -Name c8yb -Value (Get-ClientBinary)

        Function ConvertTo-JsonPipe {
            [cmdletbinding()]
            Param(
                [Parameter(
                    ValueFromPipeline = $true,
                    Position = 0
                )]
                [object[]] $InputObject
            )
            Process {
                $InputObject | ForEach-Object { ConvertTo-Json $_ -Depth 100 -Compress }
            }
        }
    }
    BeforeEach {

    }

    Context "Job limits" {
        It "stops early due to job limit being exceeded" {
            $output = @("1", "2", "3") | c8yb events get --maxJobs 1 --dry --verbose 2>&1
            $LASTEXITCODE | Should -Be 105
            $output | Should -ContainRequest "GET /event/events/1" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 1
        }

        It "stops early due to job limit being exceeded using env variable" {
            $env:C8Y_SETTINGS_DEFAULT_BATCHMAXIMUMJOBS = "2"
            $output = @("1", "2", "3") | c8yb events get --dry --verbose 2>&1
            $env:C8Y_SETTINGS_DEFAULT_BATCHMAXIMUMJOBS = ""
            $LASTEXITCODE | Should -Be 105
            $output | Should -ContainRequest "GET /event/events/1" -Total 1
            $output | Should -ContainRequest "GET /event/events/2" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }
    }


    AfterEach {
    }
}
