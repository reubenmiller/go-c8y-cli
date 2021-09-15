. $PSScriptRoot/imports.ps1 -ErrorAction SilentlyContinue -SkipSessionTest

Describe -Tag "Session" -Name "Set-Session" {
    BeforeAll {
        $EnvBackup = Get-Item "Env:C8Y*"
        foreach ($item in $EnvBackup) {
            Remove-Item ("Env:{0}" -f $item.Key)
        }
    }

    BeforeEach {
        $tmpdir = New-TemporaryDirectory
        $env:C8Y_SESSION_HOME = $tmpdir
        $env:C8Y_HOME = $tmpdir
        $settingsFile = "$tmpdir/settings.json"
    }

    It "Loads a session from a folder by name" {
        # session
        $env:C8Y_SESSION = ""
        Clear-Session

        $Session = @{
            "host" = "https://example.com"
            "settings.defaults.pageSize" = 44
        }
        $Session | ConvertTo-Json | Out-File "$tmpdir/my-session.json"

        $resp = c8y devices list --session "my-session.json" --dry --dryFormat json | ConvertFrom-Json
        $LASTEXITCODE | Should -BeExactly 0

        $resp.host | Should -BeExactly $Session.host
        $resp.query | Should -BeLike "*pageSize=44*"

        $settings = c8y settings list --session "my-session.json" | ConvertFrom-Json
        $settings.defaults.pageSize | Should -BeExactly 44
    }

    It "Loads a common preferences from the session folder automatically" {
        $env:C8Y_SESSION = ""
        Clear-Session
        $Settings = @{
            "settings.includeAll.pageSize" = 123
        }
        $Settings | ConvertTo-Json | Out-File $settingsFile
        
        $session_settings = c8y settings list | ConvertFrom-Json
        $LASTEXITCODE | Should -BeExactly 0
        $session_settings.includeAll.pageSize | Should -BeExactly 123
    }

    It "Session settings override common preferences" {
        # settings
        $Settings = @{
            "settings.defaults.pageSize" = 120
            "settings.includeAll.delayMS" = 23
        }
        $Settings | ConvertTo-Json | Out-File $settingsFile

        # session
        $env:C8Y_SESSION = "$tmpdir/my-session.json"
        $Session = @{
            "settings.defaults.pageSize" = 99
        }
        $Session | ConvertTo-Json | Out-File $env:C8Y_SESSION

        $session_settings = c8y settings list | ConvertFrom-Json
        $LASTEXITCODE | Should -BeExactly 0
        $session_settings.defaults.pageSize | Should -BeExactly 99
        $session_settings.includeAll.delayMS | Should -BeExactly 23
    }

    It "Session settings without preferences" {
        # session
        $env:C8Y_SESSION = "$tmpdir/my-session2.json"
        $Session = @{
            "settings.defaults.pageSize" = 24
        }
        $Session | ConvertTo-Json | Out-File $env:C8Y_SESSION

        $session_settings = c8y settings list | ConvertFrom-Json
        $LASTEXITCODE | Should -BeExactly 0
        $session_settings.defaults.pageSize | Should -BeExactly 24
        $session_settings.includeAll.pageSize | Should -BeExactly 2000
    }

    It "Loads a yaml session the current directory called session.yaml" {
        # session
        $sessionFile = "$tmpdir/session.yaml"
        $env:C8Y_SESSION = ""
        @"
settings:
    defaults:
        pageSize: 110
settings.includeAll.pagesize: 202
"@ | Out-File $sessionFile

        $session_settings = c8y settings list | ConvertFrom-Json
        $LASTEXITCODE | Should -BeExactly 0
        $session_settings.defaults.pageSize | Should -BeExactly 110
        $session_settings.includeAll.pageSize | Should -BeExactly 202

        $resp -like "*settings.default.pageSize: 110" | Should -HaveCount 1
        $resp -like "*settings.includeAll.pageSize: 202" | Should -HaveCount 1
    }

    AfterEach {
        Remove-Item $tmpdir -Force -Recurse -ErrorAction SilentlyContinue
    }

    AfterAll {
        # Restore env variables
        foreach ($item in $EnvBackup) {
            Set-Item -Path ("env:{0}" -f $item.Key) -Value $item.Value
        }
    }
}
