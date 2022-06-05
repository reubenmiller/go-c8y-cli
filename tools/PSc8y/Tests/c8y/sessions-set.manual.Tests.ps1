. $PSScriptRoot/../imports.ps1 -ErrorAction SilentlyContinue -SkipSessionTest

Describe -Tag "Session" -Name "sessions-set" {
    BeforeAll {
        $tmpdir = New-TemporaryDirectory

        $UserTemplate = "{userName: 'testuser_' + _.Char(6), password: _.Password() }"
        $input01 = c8y template execute --template $UserTemplate | ConvertFrom-Json
        $input02 = c8y template execute --template $UserTemplate | ConvertFrom-Json
        c8y users create --userName $input01.userName --password $input01.password
        c8y users create --userName $input02.userName --password $input02.password

        $encrypted02 = c8y sessions encryptText --text $input02.password --passphrase test --raw
        $c8yhost = (Get-Session).host -replace "https://", ""

        $Session01 = Join-Path $tmpdir "session01.json"
        $Session02 = Join-Path $tmpdir "session02.json"
        @{
            username = $input01.userName
            password = $input01.password
            host = $c8yhost
        } | ConvertTo-Json | Out-File $Session01

        @{
            username = $input02.userName
            password = $encrypted02
            host = $c8yhost
        } | ConvertTo-Json | Out-File $Session02

        $EnvBackup = Get-Item "Env:C8Y*"
        foreach ($item in $EnvBackup) {
            Remove-Item ("Env:{0}" -f $item.Key)
        }

        $env:C8Y_SESSION_HOME = $tmpdir
    }

    BeforeEach {
        
    }

    It "uses a session without environment variables (unencrypted and encrypted)" {
        # session
        $env:C8Y_SESSION = $Session01
        $env:C8Y_USERNAME | Should -BeNullOrEmpty

        # session 1
        $env:C8Y_PASSPHRASE = "test"
        $output = c8y inventory list --dry --dryFormat json
        $LASTEXITCODE | Should -BeExactly 0
        $request = $output | ConvertFrom-Json

        $request | Should -MatchObject @{
            path = "/inventory/managedObjects"
        } -Property path

        $basicauth = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($input01.userName + ":" + $input01.password))
        $request.headers.Authorization | Should -Match "Basic $basicauth"

        # session 2
        $env:C8Y_PASSPHRASE = "test"
        $output = c8y inventory list --dry --dryFormat json --session $Session02
        $LASTEXITCODE | Should -BeExactly 0
        $request = $output | ConvertFrom-Json

        $request | Should -MatchObject @{
            path = "/inventory/managedObjects"
        } -Property path

        $basicauth = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($input02.userName + ":" + $input02.password))
        $request.headers.Authorization | Should -Match "Basic $basicauth"
    }

    It "uses a session from environment variables (unencrypted and encrypted)" {
        # session
        $env:C8Y_SESSION = $null
        $env:C8Y_HOST = $c8yhost
        $env:C8Y_USERNAME = $input01.userName
        $env:C8Y_PASSWORD = $input01.password

        # session 1
        $output = c8y inventory list --dry --dryFormat json
        $LASTEXITCODE | Should -BeExactly 0
        $request = $output | ConvertFrom-Json

        $request | Should -MatchObject @{
            path = "/inventory/managedObjects"
        } -Property path

        $basicauth = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($input01.userName + ":" + $input01.password))
        $request.headers.Authorization | Should -Match "Basic $basicauth"

        # session 2
        $env:C8Y_PASSPHRASE = "test"
        $env:C8Y_SESSION = $null
        $env:C8Y_HOST = $c8yhost
        $env:C8Y_USERNAME = $input02.userName
        $env:C8Y_PASSWORD = $encrypted02

        $output = c8y inventory list --dry --dryFormat json --session $Session02
        $LASTEXITCODE | Should -BeExactly 0
        $request = $output | ConvertFrom-Json

        $request | Should -MatchObject @{
            path = "/inventory/managedObjects"
        } -Property path

        $basicauth = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($input02.userName + ":" + $input02.password))
        $request.headers.Authorization | Should -Match "Basic $basicauth"
    }

    It "fails when using invalid passphrase on an encrypted session" {
        # session
        $env:C8Y_SESSION = $null
        $env:C8Y_HOST = $c8yhost
        $env:C8Y_USERNAME = $input02.userName
        $env:C8Y_PASSWORD = $encrypted02
        $env:C8Y_PASSPHRASE = "wrong_passphrase"

        c8y inventory list --dry --dryFormat json
        $LASTEXITCODE | Should -BeExactly 108
    }


    It "combines session configuration from environment and file" {
        # session
        $Session03 = Join-Path $tmpdir "session03.json"
        Get-Content $Session01 | ConvertFrom-Json | Select-Object username, password | ConvertTo-Json | Out-File $Session03
        $env:C8Y_SESSION = $session03
        $env:C8Y_HOST = $c8yhost
        $env:C8Y_USERNAME = $null
        $env:C8Y_PASSWORD = $null

        # using C8Y_SESSION
        $output = c8y inventory list --dry --dryFormat json
        $LASTEXITCODE | Should -BeExactly 0
        $request = $output | ConvertFrom-Json

        $request | Should -MatchObject @{
            path = "/inventory/managedObjects"
        } -Property path

        $basicauth = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($input01.userName + ":" + $input01.password))
        $request.headers.Authorization | Should -Match "Basic $basicauth"
        $request.host | Should -BeExactly ("https://" + $c8yhost)

        # Use --session parameter (should override C8Y_SESSION value)
        $env:C8Y_SESSION = $session02
        $output = c8y inventory list --dry --dryFormat json --session $session03
        $LASTEXITCODE | Should -BeExactly 0
        $request = $output | ConvertFrom-Json

        $request | Should -MatchObject @{
            path = "/inventory/managedObjects"
        } -Property path

        $basicauth = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($input01.userName + ":" + $input01.password))
        $request.headers.Authorization | Should -Match "Basic $basicauth"
        $request.host | Should -BeExactly ("https://" + $c8yhost)
    }

    AfterAll {
        Clear-Session

        # Restore env variables
        foreach ($item in $EnvBackup) {
            Set-Item -Path ("env:{0}" -f $item.Key) -Value $item.Value
        }

        Remove-Item $tmpdir -Force -Recurse -ErrorAction SilentlyContinue
        $input01.userName, $input02.userName | Remove-User
    }
}
