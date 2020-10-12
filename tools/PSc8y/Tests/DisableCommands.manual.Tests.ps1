. $PSScriptRoot/imports.ps1

Describe -Name "Disable create/update/delete commands" {
    BeforeAll {
        $ciSetting = $env:C8Y_SETTINGS_CI
    }

    BeforeEach {
        $env:C8Y_SETTINGS_CI = ""
        $env:C8Y_SETTINGS_MODE_ENABLECREATE = ""
        $env:C8Y_SETTINGS_MODE_ENABLEUPDATE = ""
        $env:C8Y_SETTINGS_MODE_ENABLEDELETE = ""

        $items = New-Object System.Collections.ArrayList
    }

    It "Enables create commands" {

        $null = New-TestDevice -WhatIf
        $LASTEXITCODE | Should -Not -Be 0

        Set-ClientConsoleSetting -EnableCreateCommands

        $null = New-TestDevice -WhatIf
        $LASTEXITCODE | Should -Be 0
    }

    It "Enables update commands" {
        Set-ClientConsoleSetting -EnableCreateCommands

        $device = New-TestDevice
        $LASTEXITCODE | Should -Be 0
        $null = $items.Add($device.id)

        # updates should not work
        $device | PSc8y\Update-Device -NewName "My New Name" -WhatIf
        $LASTEXITCODE | Should -Not -Be 0

        Set-ClientConsoleSetting -EnableUpdateCommands

        # updates should work
        $device | PSc8y\Update-Device -NewName "My New Name" -WhatIf
        $LASTEXITCODE | Should -Be 0
    }

    It "Enables delete commands" {
        Set-ClientConsoleSetting -EnableCreateCommands

        $device = New-TestDevice
        $LASTEXITCODE | Should -Be 0
        $null = $items.Add($device.id)

        # delete should not work
        $device | PSc8y\Remove-Device -WhatIf
        $LASTEXITCODE | Should -Not -Be 0

        Set-ClientConsoleSetting -EnableDeleteCommands

        # delete should work
        $device | PSc8y\Remove-Device
        $LASTEXITCODE | Should -Be 0
    }

    AfterEach {
        Set-ClientConsoleSetting -EnableDeleteCommands
        foreach ($item in $items) {
            if ($item) {
                PSc8y\Remove-ManagedObject -Id $item
            }
        }
    }

    AfterAll {
        if ($ciSetting) {
            $env:C8Y_SETTINGS_CI = $ciSetting
        }
    }
}
