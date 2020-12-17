. $PSScriptRoot/imports.ps1

Describe -Name "Get-AssetParent" {
    Context "Test agent/devices" {
        BeforeAll {

            $RootGroup = New-TestDeviceGroup -Name "rootgroup"
            $SubGroup01 = New-TestDeviceGroup -Name "group01" -Type SubGroup
            $SubGroup02 = New-TestDeviceGroup -Name "group02" -Type SubGroup
            
            # Add child relationships: rootgroup -> group01 -> group02
            Add-AssetToGroup -Group $RootGroup.id -NewChildGroup $SubGroup01.id
            Add-AssetToGroup -Group $SubGroup01.id -NewChildGroup $SubGroup02.id
        }

        It "Should return nothing if the asset has no parent" {
            $Response = PSc8y\Get-AssetParent `
                -Asset $RootGroup.id
            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty
        }

        It "Should return the immediate parent asset by default" {
            $Response = PSc8y\Get-AssetParent `
                -Asset $SubGroup02.id
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.id | Should -BeExactly $SubGroup01.id
            $Response.name | Should -BeExactly $SubGroup01.name
        }

        It "Should return the root parent device" {
            $Response = PSc8y\Get-AssetParent `
                -Asset $SubGroup02.id `
                -RootParent
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.id | Should -BeExactly $RootGroup.id
            $Response.name | Should -BeExactly $RootGroup.name
        }

        It "Should return the root parent when Level is out of bounds" {
            $Response = PSc8y\Get-AssetParent `
                -Asset $SubGroup02.id `
                -Level 100
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.id | Should -BeExactly $RootGroup.id
            $Response.name | Should -BeExactly $RootGroup.name
        }

        It "Should return all parent devices as an array" {
            $Response = PSc8y\Get-AssetParent `
                -Asset $SubGroup02.id `
                -All
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response | Should -HaveCount 2
            $Response.id | Should -BeExactly @($RootGroup.id, $SubGroup01.id)
        }

        AfterAll {
            # Cleanup: Delete in reverse order because sub assets are deleted by default
            @($SubGroup02, $SubGroup01, $RootGroup) | ForEach-Object {
                if ($_.id) {
                    PSc8y\Remove-ManagedObject -Id $_.id
                }
            }
        }
    }
}
