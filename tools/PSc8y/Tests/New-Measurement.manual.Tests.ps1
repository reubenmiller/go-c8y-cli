
. $PSScriptRoot/imports.ps1

Describe -Name "New Measurement" {
    BeforeEach {
        $TestDevice = PSc8y\New-ManagedObject `
            -Name "testdevice001" `
            -Data @{
                c8y_IsDevice = @{}
            }

        $TestDevice.id | Should -Not -BeNullOrEmpty
    }
    It "Data" {
        $TestDevice.id | Should -Not -BeNullOrEmpty
        $Response = PSc8y\New-Measurement `
            -Device $TestDevice.id `
            -Time "0d" `
            -Type "ciSeria1" `
            -Data @{
                test1 = @{
                    signal1 = @{
                        value = 1.234;
                        unit = "°"
                    }
                }
            }
        $Response | Should -Not -BeNullOrEmpty
        $Response.test1.signal1.value | Should -BeExactly 1.234
        $Response.test1.signal1.unit | Should -BeExactly "°"
    }

    AfterEach {
        $TestDevice.id | Should -Not -BeNullOrEmpty
        $TestDevice.id | PSc8y\Remove-ManagedObject
    }
}
