. $PSScriptRoot/imports.ps1 -ErrorAction SilentlyContinue -SkipSessionTest

Describe -Tag "Session" -Name "Login and Session Tests" {
    BeforeAll {
        $EnvBackup = Get-Item "Env:C8Y*"
        $SessionBackup = Get-Session
        $EnvBackupHash = @{}
        foreach ($item in $EnvBackup) {
            $EnvBackupHash[$item.Key] = $item.Value
            Remove-Item ("Env:{0}" -f $item.Key)
        }
    }

    BeforeEach {
        $tmpdir = New-TemporaryDirectory
        $env:C8Y_SESSION_HOME = $tmpdir
        $settingsFile = "$tmpdir/settings.json"
    }

    It -Skip "Login using OAUTH2 strategy without a session file" {
        # session
        $env:C8Y_SESSION = ""
        c8y sessions set --session "my-session"
        
        $resp = c8y devices list --verbose --dry --session "my-session"
        $LASTEXITCODE | Should -BeExactly 0
    }

    It -Skip "Logs in with TFA (TOTP) and OAUTH_INTERNAL enabled" {

    }

    It "Logs in using BASIC_AUTHORIZATION if the environment variables are set" {
        $env:C8Y_HOST = $EnvBackupHash["C8Y_HOST"]
        $env:C8Y_TENANT = $EnvBackupHash["C8Y_TENANT"]
        $env:C8Y_USER = $EnvBackupHash["C8Y_USER"]
        $env:C8Y_PASSWORD = $EnvBackupHash["C8Y_PASSWORD"]
        $env:C8Y_PASSPHRASE = $EnvBackupHash["C8Y_PASSPHRASE"]

        $resp = c8y devices list --raw
        $LASTEXITCODE | Should -BeExactly 0
        $results = $resp | ConvertFrom-Json
        $results | Should -Not -BeNullOrEmpty

        $env:C8Y_PASSWORD = "wrong password"
        $resp = c8y devices list
        $LASTEXITCODE | Should -Not -BeExactly 0
    }

    It -Skip "Uses encryption to store passwords and authorization token" {

    }

    It -Skip "Prompt for c8y password again if the passphrase has been lost" {
        
    }

    It "Does not require tenant name" {
        $env:C8Y_HOST = $EnvBackupHash["C8Y_HOST"]
        $env:C8Y_USER = $EnvBackupHash["C8Y_USER"]
        $env:C8Y_PASSWORD = $EnvBackupHash["C8Y_PASSWORD"]
        $env:C8Y_PASSPHRASE = $EnvBackupHash["C8Y_PASSPHRASE"]

        $resp = c8y devices list --raw
        $LASTEXITCODE | Should -BeExactly 0
        $results = $resp | ConvertFrom-Json
        $results | Should -Not -BeNullOrEmpty
    }

    It "Saves tenant name in the session file" {
        $env:C8Y_PASSPHRASE = "TestPassword"
        $SessionFile = Join-Path -Path $tmpdir -ChildPath "session.json"
        $env:C8Y_SESSION = $SessionFile

        $passwordText = c8y sessions decryptText --text $SessionBackup.password --passphrase $env:C8Y_PASSPHRASE
        $LASTEXITCODE | Should -BeExactly 0

        $passwordText | Should -Not -Match "^{encrypted}.+$"

        $SessionBefore = @{
            host = $SessionBackup.host
            username = $SessionBackup.username
            password = $passwordText
            settings = @{
                encryption = @{
                    enabled = $true
                }
            }
        }
        $SessionBefore | ConvertTo-Json | Out-File $SessionFile

        # Start login
        c8y sessions set --session $SessionFile
        $LASTEXITCODE | Should -BeExactly 0
        
        $SessionAfterLogin = Get-Content $SessionFile | ConvertFrom-Json
        $SessionAfterLogin.password | Should -Match "^{encrypted}.+$"
        $SessionAfterLogin.host | Should -BeExactly $SessionBefore.host
        $SessionAfterLogin.tenant | Should -Match "^t\d+$"
        $SessionAfterLogin.username | Should -BeExactly $SessionBefore.username
        $SessionAfterLogin.password | Should -Not -Be $passwordText

        # Only if OAUTH2 is being used
        if ($SessionAfterLogin.token) {
            $SessionAfterLogin.token | Should -Match "^{encrypted}.+$"
        }

        $SessionAfterLogin.'$schema' | Should -Match "^https://.+"

        # Tenant is optional
        if ($SessionAfterLogin.tenant) {
            $SessionAfterLogin.tenant | Should -Match "t\d+"
        }
        $env:C8Y_PASSPHRASE = ""
    }

    It -Skip "Switches between two login types OAUTH and BASIC_AUTH" {

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
