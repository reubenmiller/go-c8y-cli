. $PSScriptRoot/imports.ps1

Describe -Name "Encoding" {
    Context "Test device" {
        BeforeAll {
            $TestDevice = PSc8y\New-TestAgent
        }

        It "Create alarm with extended character set" {
            $Text = "Test Alarm äöüßáàæåāø¡µ~∫√ç≈≈¥å∂ƒ©ªº∆@œæ•πø⁄¨Ω†∑«¡¶¢[[]|{}≠¿"
            $Response = PSc8y\New-Alarm `
                -Device $TestDevice.id `
                -Type c8y_TestAlarm `
                -Time "-0s" `
                -Text $Text `
                -Severity MAJOR
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.text | Should -BeExactly $Text
        }

        It "Create event with extended character set" {
            $Text = "Test Event äöüßáàæåāø¡µ~∫√ç≈≈¥å∂ƒ©ªº∆@œæ•πø⁄¨Ω†∑«¡¶¢[[]|{}≠¿"
            $Response = PSc8y\New-Event `
                -Device $TestDevice.id `
                -Type c8y_TestEvent `
                -Time "-0s" `
                -Text $Text
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.text | Should -BeExactly $Text
        }

        It "Create operation with extended character set" {
            $Text = "Test Operation äöüßáàæåāø¡µ~∫√ç≈≈¥å∂ƒ©ªº∆@œæ•πø⁄¨Ω†∑«¡¶¢[[]|{}≠¿"
            $Response = PSc8y\New-Operation `
                -Device $TestDevice.id `
                -Description $Text `
                -Data @{
                    c8y_TestOperation = @{
                        parameters = @{
                            value1 = $Text
                        }
                    }
                }
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response.description | Should -BeExactly $Text
            $Response.c8y_TestOperation.parameters.value1 | Should -BeExactly $Text
        }

        It "Create measurements with extended character sets in value fragment" {
            $TestDevice.id | Should -Not -BeNullOrEmpty
            $Response = PSc8y\New-Measurement `
                -Device $TestDevice.id `
                -Time "0d" `
                -Type "ciSeria1" `
                -Data @{
                    "ÄnderungZahler" = @{
                        "ö1" = @{
                            value = 2;
                            unit = "°"
                        }
                    }
                }
            $Response | Should -Not -BeNullOrEmpty
            $Response.ÄnderungZahler.ö1.value | Should -BeExactly 2
            $Response.ÄnderungZahler.ö1.unit | Should -BeExactly "°"
        }

        It "Create measurement and return raw json" {
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
                } `
                -Raw
            $Response | Should -Not -BeNullOrEmpty

            $ResponseJSON = $Response | ConvertTo-Json -Depth 100 -Compress
            $ResponseJSON | Should -BeLike '*"value":1.234*'
            $ResponseJSON | Should -BeLike '*"unit":"°"*'
        }

        AfterAll {
            if ($TestDevice.id) {
                PSc8y\Remove-ManagedObject -Id $TestDevice.id
            }
        }
    }
}
