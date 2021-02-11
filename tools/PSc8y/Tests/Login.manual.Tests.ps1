. $PSScriptRoot/imports.ps1 -ErrorAction SilentlyContinue -SkipSessionTest

Describe -Tag "Session" -Name "Login and Session Tests" {
    BeforeAll {
        . "$PSScriptRoot/New-TemporaryDirectory.ps1"
        $EnvBackup = Get-Item "Env:C8Y*"
        $SessionBackup = Get-Session
        $EnvBackupHash = @{}
        foreach ($item in $EnvBackup) {
            $EnvBackupHash[$item.Name] = $item.Value
            Remove-Item ("Env:{0}" -f $item.Name)
        }
    }

    BeforeEach {
        $tmpdir = New-TemporaryDirectory
        $env:C8Y_SESSION_HOME = $tmpdir
        $env:C8Y_USE_ENVIRONMENT = ""
        $settingsFile = "$tmpdir/settings.json"

        $c8y = Get-ClientBinary
    }

    It -Skip "Login using OAUTH2 strategy without a session file" {
        # session
        $env:C8Y_SESSION = ""
        & $c8y sessions login
        
        $resp = & $c8y devices list --verbose --dry --session "my-session" 2>&1
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

        $resp = & $c8y devices list --raw
        $LASTEXITCODE | Should -BeExactly 0
        $results = $resp | ConvertFrom-Json
        $results | Should -Not -BeNullOrEmpty

        $env:C8Y_PASSWORD = "wrong password"
        $resp = & $c8y devices list
        $LASTEXITCODE | Should -Not -BeExactly 0
    }

    It -Skip "Uses encryption to store passwords and authorization cookies" {

    }

    It -Skip "Prompt for c8y password again if the passphrase has been lost" {
        
    }

    It "Does not require tenant name" {
        $env:C8Y_HOST = $EnvBackupHash["C8Y_HOST"]
        $env:C8Y_USER = $EnvBackupHash["C8Y_USER"]
        $env:C8Y_PASSWORD = $EnvBackupHash["C8Y_PASSWORD"]
        $env:C8Y_PASSPHRASE = $EnvBackupHash["C8Y_PASSPHRASE"]

        $resp = & $c8y devices list --raw
        $LASTEXITCODE | Should -BeExactly 0
        $results = $resp | ConvertFrom-Json
        $results | Should -Not -BeNullOrEmpty
    }

    It "Saves tenant name in the session file" {
        $env:C8Y_PASSPHRASE = "TestPassword"
        $SessionFile = Join-Path -Path $tmpdir -ChildPath "session.json"
        $env:C8Y_SESSION = $SessionFile

        $passwordText = & $c8y sessions decryptText --text $SessionBackup.password --passphrase $env:C8Y_PASSPHRASE
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
        & $c8y sessions login
        $LASTEXITCODE | Should -BeExactly 0
        
        $SessionAfterLogin = Get-Content $SessionFile | ConvertFrom-Json
        $SessionAfterLogin.password | Should -Match "^{encrypted}.+$"
        $SessionAfterLogin.host | Should -BeExactly $SessionBefore.host
        $SessionAfterLogin.tenant | Should -Match "^t\d+$"
        $SessionAfterLogin.username | Should -BeExactly $SessionBefore.username
        $SessionAfterLogin.password | Should -Not -Be $passwordText
        $SessionAfterLogin.credential | Should -Not -BeNullOrEmpty

        # Only if OAUTH2 is being used
        if ($SessionAfterLogin.credential.cookies.0) {
            $SessionAfterLogin.credential.cookies.0 | Should -Match "^{encrypted}.+$"
        }

        if ($SessionAfterLogin.credential.cookies.1) {
            $SessionAfterLogin.credential.cookies.1 | Should -Match "^{encrypted}.+$"
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
            Set-Item -Path ("env:{0}" -f $item.Name) -Value $item.Value
        }
    }
}
