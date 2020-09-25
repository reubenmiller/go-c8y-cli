. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceGroupCollection" {
    Context "Device groups with spaces in their names" {
        BeforeAll {
            $RandomPart = New-RandomString
            $Group01 = New-TestDeviceGroup -Name "My Custom Group $RandomPart" -Type Group
            $Group02 = New-TestDeviceGroup -Name "My Custom Group $RandomPart" -Type SubGroup
        }

        It "Find device groups by name" {
            $Response = PSc8y\Get-DeviceGroupCollection -Name "*My Custom Group ${RandomPart}*" -PageSize 5
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty

            $Response.Count | Should -BeExactly 2
        }

        It "Find sub (nested) device groups" {
            [array] $Response = PSc8y\Get-DeviceGroupCollection -Name "*My Custom Group ${RandomPart}*" -ExcludeRootGroup -PageSize 5
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty

            $Response.Count | Should -BeExactly 1
            $Response[0].name | Should -BeExactly $Group02.name
        }

        It "Find root device groups" {
            [array] $Response = PSc8y\Get-DeviceGroupCollection -Name "*My Custom Group ${RandomPart}*" -Type "c8y_DeviceGroup" -PageSize 5
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty

            $Response.Count | Should -BeExactly 1
            $Response[0].name | Should -BeExactly $Group01.name
        }

        AfterAll {
            $null = Remove-ManagedObject -Id $Group01.id -ErrorAction SilentlyContinue 2>&1
            $null = Remove-ManagedObject -Id $Group02.id -ErrorAction SilentlyContinue 2>&1
        }
    }
}
