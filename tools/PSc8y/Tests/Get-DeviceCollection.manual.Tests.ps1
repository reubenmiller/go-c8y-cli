. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceCollection" {
    Context "Devices with spaces in their names" {
        BeforeAll {
            $RandomPart = New-RandomString
            $Device01 = New-TestDevice -Name "My Custom Device $RandomPart"
            $Device02 = New-TestDevice -Name "My Custom Device $RandomPart"
        }

        It "Find devices by name" {
            $Response = PSc8y\Get-DeviceCollection -Name "*My Custom Device ${RandomPart}*" -PageSize 5
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty

            $Response.Count | Should -BeExactly 2
        }

        It "Returns all devices using includeAll with WhatIf" {
            $Response = PSc8y\Get-DeviceCollection -Verbose -IncludeAll -WhatIf
            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty
        }

        AfterAll {
            $null = Remove-ManagedObject -Id $Device01.id -ErrorAction SilentlyContinue 2>&1
            $null = Remove-ManagedObject -Id $Device02.id -ErrorAction SilentlyContinue 2>&1
        }
    }
}
