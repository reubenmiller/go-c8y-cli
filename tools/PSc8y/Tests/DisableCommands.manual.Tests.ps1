. $PSScriptRoot/imports.ps1

Describe -Name "Disable create/update/delete commands" {
    BeforeAll {
        $backupEnvSettings = @{
            CI = $env:CI
            C8Y_SETTINGS_CI = $env:C8Y_SETTINGS_CI
            C8Y_MODE = $env:C8Y_MODE
        }
    }

    BeforeEach {
        $env:CI = "false"
        $env:C8Y_SETTINGS_CI = "false"
        $env:C8Y_MODE = "prod"

        $items = New-Object System.Collections.ArrayList
    }

    It "Enables create commands" {

        $null = New-TestDevice
        $LASTEXITCODE | Should -Not -Be 0

        Set-ClientConsoleSetting -Mode dev

        $device = New-TestDevice
        $items.Add($device.id)
        $LASTEXITCODE | Should -Be 0
    }

    It "Enables update commands" {
        Set-ClientConsoleSetting -Mode prod

        $device = New-TestDevice
        $LASTEXITCODE | Should -Be 0
        $null = $items.Add($device.id)

        # updates should not work
        $device | PSc8y\Update-Device -NewName "My New Name"
        $LASTEXITCODE | Should -Not -Be 0

        Set-ClientConsoleSetting -Mode qual

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
        Set-ClientConsoleSetting -Mode qual

        $device = New-TestDevice
        $LASTEXITCODE | Should -Be 0
        $null = $items.Add($device.id)

        # delete should not work
        $device | PSc8y\Remove-Device
        $LASTEXITCODE | Should -Not -Be 0

        Set-ClientConsoleSetting -Mode dev

        # delete should work
        $device | PSc8y\Remove-Device
        $LASTEXITCODE | Should -Be 0
    }

    AfterEach {
        foreach ($item in $items) {
            if ($item) {
                PSc8y\Remove-ManagedObject -Id $item -SessionMode dev
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
