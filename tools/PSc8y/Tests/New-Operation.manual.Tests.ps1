. $PSScriptRoot/imports.ps1

Describe -Name "New-Operation manual tests" {
    BeforeEach {
        $device = New-TestAgent

    }

    It "Create operation with nested values" {
        $data = @{
            c8y_Nested = @{
                type1 = @{
                    names = @{
                        sorted = @{
                            values = @(1,2,3)
                        }
                    }
                }
            }
        }
        $Response = PSc8y\New-Operation -Device $device.id -Description "Restart device" -Data $data
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response.c8y_Nested.type1.names.sorted.values | Should -BeExactly @(1,2,3)
    }

    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}

