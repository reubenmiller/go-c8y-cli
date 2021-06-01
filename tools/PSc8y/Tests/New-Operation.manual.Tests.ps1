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

    It "Create operation using device name" {
        $data = @{
            c8y_Restart = @{}
        }
        $Response = PSc8y\New-Operation -Device $device.name -Description "Restart device" -Data $data
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response.deviceId | Should -BeExactly $device.id
    }

    It "Override piped input with explict device" {
        $inputdata = @{
            deviceId = "12345"
            c8y_Restart = @{}
        } | ConvertTo-Json -Compress
        $output = $inputdata | c8y operations create --dry --template "{v: input.value}" --device $device.name --dry --dryFormat json
        $LASTEXITCODE | Should -Be 0
        $request = $output | ConvertFrom-Json

        $request.body.deviceId | Should -BeExactly $device.id
        $request.body.v | Should -MatchObject @{
            c8y_Restart = @{}
            deviceId = "12345"
        }
    }

    

    AfterEach {
        Remove-ManagedObject -Id $device.id

    }
}

