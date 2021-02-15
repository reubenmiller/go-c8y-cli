. $PSScriptRoot/../imports.ps1

Describe -Name "c8y activitylog" {
    BeforeEach {
        $backupEnvSettings = @{
            C8Y_SETTINGS_ACTIVITYLOG_PATH = $Env:C8Y_SETTINGS_ACTIVITYLOG_PATH
            C8Y_SETTINGS_ACTIVITYLOG_ENABLED = $Env:C8Y_SETTINGS_ACTIVITYLOG_ENABLED
            C8Y_SETTINGS_ACTIVITYLOG_METHODFILTER = $Env:C8Y_SETTINGS_ACTIVITYLOG_METHODFILTER
        }

        $TempFile = New-TemporaryFile
        Remove-Item $TempFile
        $ActiveLogDir = New-Item -ItemType Directory -Path $TempFile
        $Env:C8Y_SETTINGS_ACTIVITYLOG_PATH = $ActiveLogDir.FullName
        $Env:C8Y_SETTINGS_ACTIVITYLOG_ENABLED = "true"
        $Env:C8Y_SETTINGS_ACTIVITYLOG_METHODFILTER = ""
    }

    Context "defaults" {
        It "logs commands and requests" {
            $null = c8y inventory create --name "myLoggedDevice" | c8y inventory delete
            $LASTEXITCODE | Should -Be 0
            $logs = Get-ChildItem $ActiveLogDir.FullName -Filter "*.json"
            $logs | Should -HaveCount 1
            $logs | should -FileContentMatch "POST"
            $logs | should -FileContentMatch "DELETE"
        }

        It "skips logging when disabled" {
            $Env:C8Y_SETTINGS_ACTIVITYLOG_ENABLED = "false"
            $null = c8y inventory create --name "myLoggedDevice" --noLog | c8y inventory delete
            $LASTEXITCODE | Should -Be 0
            $logs = Get-ChildItem $ActiveLogDir.FullName -Filter "*.json"
            $logs | Should -HaveCount 0
        }

        It "skips logging when noLog is used" {
            $null = c8y inventory create --name "myLoggedDevice" --noLog | c8y inventory delete
            $LASTEXITCODE | Should -Be 0
            $logs = Get-ChildItem $ActiveLogDir.FullName -Filter "*.json"
            $logs | Should -HaveCount 1
            $logs | should -Not -FileContentMatchExactly "POST"
            $logs | should -FileContentMatchExactly "DELETE"
        }

        It "skips specific rest request methods" {
            $Env:C8Y_SETTINGS_ACTIVITYLOG_METHODFILTER = "POST PUT"
            $null = c8y inventory create --name "myLoggedDevice" `
                | c8y inventory update --newName "myUpdatedLoggedDevice" `
                | c8y inventory delete
            $LASTEXITCODE | Should -Be 0
            $logs = Get-ChildItem $ActiveLogDir.FullName -Filter "*.json"
            $logs | Should -HaveCount 1
            $logs | should -FileContentMatchExactly "POST"
            $logs | should -FileContentMatchExactly "PUT"
            $logs | should -Not -FileContentMatchExactly "DELETE"
        }
    }

    AfterEach {
        if ($ActiveLogDir -and (Test-Path $ActiveLogDir)) {
            Remove-Item $ActiveLogDir -Recurse -Force
        }

        if ($backupEnvSettings) {
            foreach ($name in $backupEnvSettings.Keys) {
                if ($null -ne $name) {
                    [environment]::SetEnvironmentVariable($name, $backupEnvSettings[$name], "process")
                }
            }
        }
    }
}
