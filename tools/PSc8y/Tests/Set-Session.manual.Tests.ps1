. $PSScriptRoot/imports.ps1 -ErrorAction SilentlyContinue -SkipSessionTest

Describe -Tag "Session" -Name "Set-Session" {
    BeforeAll {
        . "$PSScriptRoot/New-TemporaryDirectory.ps1"
        $sessionHomeBackup = $env:C8Y_SESSION_HOME
        $sessionBackup = $env:C8Y_SESSION
        $sessionBackupHost = $env:C8Y_HOST
        $sessionBackupTenant = $env:C8Y_TENANT
        $sessionBackupUser = $env:C8Y_USER
        $sessionBackupPassword = $env:C8Y_PASSWORD
        $sessionBackupUseEnv = $env:C8Y_USE_ENVIRONMENT
    }

    BeforeEach {
        $tmpdir = New-TemporaryDirectory
        $env:C8Y_SESSION_HOME = $tmpdir
        $env:C8Y_USE_ENVIRONMENT = ""
        $settingsFile = "$tmpdir/settings.json"
    }

    It "Loads a session from a folder by name" {
        # session
        $env:C8Y_SESSION = ""
        $Session = @{
            "host" = "https://example.com"
            "settings.default.pageSize" = 44
        }
        $Session | ConvertTo-Json | Out-File "$tmpdir/my-session.json"

        $c8y = Get-ClientBinary
        $resp = & $c8y devices list --verbose --dry --session "my-session" 2>&1
        $LASTEXITCODE | Should -BeExactly 0

        $resp -like "*https://example.com/inventory/managedObjects*" | Should -HaveCount 1
        $resp -like "*settings.default.pageSize: 44" | Should -HaveCount 1
    }

    It "Loads a common preferences from the session folder automatically" {
        $env:C8Y_SESSION = ""
        $Settings = @{
            "settings.includeAll.pageSize" = 123
        }
        $Settings | ConvertTo-Json | Out-File $settingsFile

        $c8y = Get-ClientBinary
        $resp = & $c8y version -v 2>&1
        $LASTEXITCODE | Should -BeExactly 0

        $resp -like "*settings.includeAll.pageSize: 123" | Should -HaveCount 1
    }

    It "Session settings override common preferences" {
        # settings
        $Settings = @{
            "settings.default.pageSize" = 120
            "settings.includeAll.delayMS" = 23
        }
        $Settings | ConvertTo-Json | Out-File $settingsFile

        # session
        $env:C8Y_SESSION = "$tmpdir/my-session.json"
        $Session = @{
            "settings.default.pageSize" = 99
        }
        $Session | ConvertTo-Json | Out-File $env:C8Y_SESSION

        $c8y = Get-ClientBinary
        $resp = & $c8y version -v 2>&1
        $LASTEXITCODE | Should -BeExactly 0

        $resp -like "*settings.default.pageSize: 99" | Should -HaveCount 1
        $resp -like "*settings.includeAll.delayMS: 23" | Should -HaveCount 1
    }

    It "Session settings without preferences" {
        # session
        $env:C8Y_SESSION = "$tmpdir/my-session2.json"
        $Session = @{
            "settings.default.pageSize" = 24
        }
        $Session | ConvertTo-Json | Out-File $env:C8Y_SESSION

        $c8y = Get-ClientBinary
        $resp = & $c8y version -v 2>&1
        $LASTEXITCODE | Should -BeExactly 0

        $resp -like "*settings.default.pageSize: 24" | Should -HaveCount 1
        $resp -like "*settings.includeAll.pageSize: 2000" | Should -HaveCount 1
    }

    It "Loads a yaml session the current directory called session.yaml" {
        # session
        $sessionFile = "$tmpdir/session.yaml"
        $env:C8Y_SESSION = ""
        @"
settings:
    default:
        pageSize: 110
settings.includeAll.pagesize: 202
"@ | Out-File $sessionFile

        $c8y = Get-ClientBinary
        $resp = & $c8y version --verbose 2>&1
        $LASTEXITCODE | Should -BeExactly 0

        $resp -like "*settings.default.pageSize: 110" | Should -HaveCount 1
        $resp -like "*settings.includeAll.pageSize: 202" | Should -HaveCount 1
    }

    AfterEach {
        Remove-Item $tmpdir -Force -Recurse -ErrorAction SilentlyContinue
    }

    AfterAll {
        $env:C8Y_SESSION_HOME = $sessionHomeBackup
        $env:C8Y_SESSION = $sessionBackup

        $env:C8Y_HOST = $sessionBackupHost
        $env:C8Y_TENANT = $sessionBackupTenant
        $env:C8Y_USER = $sessionBackupUser
        $env:C8Y_PASSWORD = $sessionBackupPassword
        $env:C8Y_USE_ENVIRONMENT = $sessionBackupUseEnv
    }
}
