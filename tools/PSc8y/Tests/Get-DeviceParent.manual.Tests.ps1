. $PSScriptRoot/imports.ps1

Describe -Name "Get-DeviceParent" {
    Context "Test agent/devices" {
        $Agent = PSc8y\New-TestAgent -Name "agent"
        $ChildDevice01 = PSc8y\New-TestDevice -Name "child01"
        $ChildDevice02 = PSc8y\New-TestDevice -Name "child02"

        # Add child relationships: Agent -> ChildDevice01 -> ChildDevice02
        New-ChildDeviceReference -Device $Agent.id -NewChild $ChildDevice01.id
        New-ChildDeviceReference -Device $ChildDevice01.id -NewChild $ChildDevice02.id

        It "Should return nothing if the device has no parent" {
            $Response = PSc8y\Get-DeviceParent `
                -Device $Agent.id
            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty
        }

        It "Should return the immediate parent device by default" {
            $Response = PSc8y\Get-DeviceParent `
                -Device $ChildDevice02.id
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.id | Should -BeExactly $ChildDevice01.id
            $Response.name | Should -BeExactly $ChildDevice01.name
        }

        It "Should return the immediate parent device using device pipeline devices" {
            $Response = PSc8y\Get-DeviceCollection -Name $ChildDevice01.name |
                PSc8y\Get-DeviceParent
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.id | Should -BeExactly $Agent.id
            $Response.name | Should -BeExactly $Agent.name
        }

        It "Should return the immediate parent device using just the device's name" {
            $Response = PSc8y\Get-DeviceParent `
                -Device $ChildDevice02.name
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.id | Should -BeExactly $ChildDevice01.id
            $Response.name | Should -BeExactly $ChildDevice01.name
        }

        It "Should return the root parent device" {
            $Response = PSc8y\Get-DeviceParent `
                -Device $ChildDevice02.id `
                -RootParent
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.id | Should -BeExactly $Agent.id
            $Response.name | Should -BeExactly $Agent.name
        }

        It "Should return the root parent when Level is out of bounds" {
            $Response = PSc8y\Get-DeviceParent `
                -Device $ChildDevice02.id `
                -Level 100
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.id | Should -BeExactly $Agent.id
            $Response.name | Should -BeExactly $Agent.name
        }

        It "Should return all parent devices as an array" {
            $Response = PSc8y\Get-DeviceParent `
                -Device $ChildDevice02.id `
                -All
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response | Should -HaveCount 2
            $Response.id | Should -BeExactly @($Agent.id, $ChildDevice01.id)
        }

        # Cleanup
        @($Agent, $ChildDevice01, $ChildDevice02) | ForEach-Object {
            if ($_.id) {
                PSc8y\Remove-ManagedObject -Id $_.id
            }
        }
    }
}
