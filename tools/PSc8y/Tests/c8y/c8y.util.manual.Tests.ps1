. $PSScriptRoot/../imports.ps1

Describe -Name "c8y util" {
    BeforeAll {
        $InputValues = @(1..10)
    }

    Context "repeat" {        
        It "outputs at least 1 line when no input is give" {
            $output = c8y util repeat
            $LASTEXITCODE | Should -Be 0
            $output | Should -HaveCount 1
            $output | Should -BeExactly " "
        }

        It "repeats piped input" {
            $output = "1" | c8y util repeat 5
            $LASTEXITCODE | Should -Be 0
            $output | Select-Object -Unique | Should -HaveCount 1
            $output | Should -HaveCount 5
        }

        It "uses a default format string which includes the input" {
            "1" `
            | c8y util repeat --times 2 `
            | Out-String | Should -BeExactly "1`n1`n"
            $LASTEXITCODE | Should -Be 0
        }

        It "uses a custom format string" {
            "1" `
            | c8y util repeat --times 2 --format "device%s" `
            | Out-String | Should -BeExactly "device1`ndevice1`n"
            $LASTEXITCODE | Should -Be 0
        }

        It "uses a custom format string with the modulus row index" {
            "1" `
            | c8y util repeat --times 2 --format "device%s-%s" `
            | Out-String | Should -BeExactly "device1-1`ndevice1-2`n"
            $LASTEXITCODE | Should -Be 0

            @("one", "two") `
            | c8y util repeat --times 2 --format "device%s-%s" `
            | Out-String | Should -BeExactly "deviceone-1`ndeviceone-2`ndevicetwo-1`ndevicetwo-2`n"
            $LASTEXITCODE | Should -Be 0
        }

        It "uses a custom format string with the current line count" {
            "1" `
            | c8y util repeat --times 2 --format "device%s-%s" --useLineCount `
            | Out-String | Should -BeExactly "device1-1`ndevice1-2`n"
            $LASTEXITCODE | Should -Be 0

            @("one", "two") `
            | c8y util repeat --times 2 --format "device%s-%s" --useLineCount `
            | Out-String | Should -BeExactly "deviceone-1`ndeviceone-2`ndevicetwo-3`ndevicetwo-4`n"
            $LASTEXITCODE | Should -Be 0
        }

        It "skips n lines" {
            $InputValues `
            | c8y util repeat --skip 1 `
            | Out-String | Should -BeExactly (($InputValues[1..9] -join "`n") + "`n")
            $LASTEXITCODE | Should -Be 0
        }

        It "prints first n lines" {
            $InputValues `
            | c8y util repeat --first 5 `
            | Out-String | Should -BeExactly (($InputValues[0..4] -join "`n") + "`n")
            $LASTEXITCODE | Should -Be 0
        }

        It "prints first n lines after skipping first n lines" {
            $InputValues `
            | c8y util repeat --first 5 --skip 2 `
            | Out-String | Should -BeExactly (($InputValues[2..6] -join "`n") + "`n")
            $LASTEXITCODE | Should -Be 0

            $InputValues `
            | c8y util repeat --first 100 --skip 2 `
            | Out-String | Should -BeExactly (($InputValues[2..9] -join "`n") + "`n")
            $LASTEXITCODE | Should -Be 0
        }

        It "combines multiple commands" {
            $expected = @(
                "device-001",
                "device-002",
                "device-003",
                "device-004",
                "device-005"
            )
            "device" `
            | c8y util repeat --format "%s" --times 5 `
            | c8y util repeat --format "%s-%03s" --useLineCount `
            | Out-String | Should -BeExactly (($expected -join "`n") + "`n")
            $LASTEXITCODE | Should -Be 0
        }

        It "combines multiple commands" {
            $expected = @(
                "device-001",
                "device-001",
                "device-001",
                "device-001",
                "device-001"
            )
            "device" `
            | c8y util repeat --format "%s" --times 5 `
            | c8y util repeat --format "%s-%03s" `
            | Out-String | Should -BeExactly (($expected -join "`n") + "`n")
            $LASTEXITCODE | Should -Be 0
        }
    }

    Context "util show" {
        It "process json line input" {
            @('{ "id": "1", "name": "device01" }') `
            | c8y util show --select id,name --output csv `
            | Should -BeExactly "1,device01"
        }

        It "filters json input" {
            @('non value') `
            | c8y util show --select id,name --output csv `
            | Should -BeExactly $null
        }
    }
}
