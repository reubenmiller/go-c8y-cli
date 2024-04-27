. $PSScriptRoot/imports.ps1

Describe -Name "Disable create/update/delete commands" {
    BeforeAll {
        $backupEnvSettings = @{
            CI = $env:CI
            C8Y_SETTINGS_CI = $env:C8Y_SETTINGS_CI
            C8Y_SETTINGS_MODE_ENABLECREATE = $env:C8Y_SETTINGS_MODE_ENABLECREATE
            C8Y_SETTINGS_MODE_ENABLEUPDATE = $env:C8Y_SETTINGS_MODE_ENABLEUPDATE
            C8Y_SETTINGS_MODE_ENABLEDELETE = $env:C8Y_SETTINGS_MODE_ENABLEDELETE
        }
    }

    BeforeEach {
        $env:CI = "false"
        $env:C8Y_SETTINGS_CI = "false"
        $env:C8Y_SETTINGS_MODE_ENABLECREATE = "false"
        $env:C8Y_SETTINGS_MODE_ENABLEUPDATE = "false"
        $env:C8Y_SETTINGS_MODE_ENABLEDELETE = "false"

        $items = New-Object System.Collections.ArrayList
    }

    It "Enables create commands" {

        $null = New-TestDevice
        $LASTEXITCODE | Should -Not -Be 0

        Set-ClientConsoleSetting -EnableCreateCommands

        $device = New-TestDevice
        $items.Add($device.id)
        $LASTEXITCODE | Should -Be 0
    }

    It "Enables update commands" {
        Set-ClientConsoleSetting -EnableCreateCommands

        $device = New-TestDevice
        $LASTEXITCODE | Should -Be 0
        $null = $items.Add($device.id)

        # updates should not work
        $device | PSc8y\Update-Device -NewName "My New Name"
        $LASTEXITCODE | Should -Not -Be 0

        Set-ClientConsoleSetting -EnableUpdateCommands

        # updates should work
        $device | PSc8y\Update-Device -NewName "My New Name"
        $LASTEXITCODE | Should -Be 0
    }

    It "Show an error to the user if the action is not allowed" {
        # updates should not work
        $output = $( $response = PSc8y\New-Device -Name "My New Name" ) 2>&1
        $LASTEXITCODE | Should -Not -Be 0
        $response | Should -BeNullOrEmpty
        $output[-1] | Should -Match "create mode is disabled"
    }

    It "Enables delete commands" {
        Set-ClientConsoleSetting -EnableCreateCommands

        $device = New-TestDevice
        $LASTEXITCODE | Should -Be 0
        $null = $items.Add($device.id)

        # delete should not work
        $device | PSc8y\Remove-Device
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
        if ($backupEnvSettings) {
            foreach ($name in $backupEnvSettings.Keys) {
                if ($null -ne $name) {
                    [environment]::SetEnvironmentVariable($name, $backupEnvSettings[$name], "process")
                }
            }
        }
    }
}
