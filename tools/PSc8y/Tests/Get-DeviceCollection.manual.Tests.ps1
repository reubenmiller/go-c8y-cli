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

        It "Find devices by name and sort by creation time (descending)" {
            $Response = PSc8y\Get-DeviceCollection -Name "*My Custom Device ${RandomPart}*" -OrderBy "creationTime desc" -PageSize 5
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.Count | Should -BeExactly 2
            $Response[0].id | Should -BeExactly $Device02.id
            $Response[1].id | Should -BeExactly $Device01.id
        }

        It "Find devices by name and sort by creation time (ascending)" {
            $Response = PSc8y\Get-DeviceCollection -Name "*My Custom Device ${RandomPart}*" -OrderBy "creationTime asc" -PageSize 5
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.Count | Should -BeExactly 2
            $Response[0].id | Should -BeExactly $Device01.id
            $Response[1].id | Should -BeExactly $Device02.id
        }

        It "Returns all devices using includeAll with WhatIf" {
            $options = @{
                IncludeAll = $true
                WhatIf = $true
                Verbose = $true
                InformationVariable = "requestInfo"
                ErrorVariable = "ErrorMessages"
            }
            $Response = PSc8y\Get-DeviceCollection @options
            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty
            $ErrorMessages | Should -Not -BeNullOrEmpty
            $requestInfo | Should -Not -BeNullOrEmpty
            $ErrorMessages | Out-String | Should -match "Using inventory optimized query"
        }

        It "Returns all devices using includeAll and verbose" {
            $overview = PSc8y\Get-DeviceCollection -PageSize 1 -WithTotalPages
            $Total = $overview.statistics.totalPages

            $devices = PSc8y\Get-DeviceCollection -Verbose -IncludeAll
            $LASTEXITCODE | Should -Be 0
            $devices | Should -Not -BeNullOrEmpty
            $devices.Count | Should -BeGreaterOrEqual ($Total * 0.7)
        }

        AfterAll {
            $null = Remove-ManagedObject -Id $Device01.id -ErrorAction SilentlyContinue 2>&1
            $null = Remove-ManagedObject -Id $Device02.id -ErrorAction SilentlyContinue 2>&1
        }
    }
}
