. $PSScriptRoot/imports.ps1

Describe -Name "Reset-UserPassword" {
    BeforeEach {
        $User = PSc8y\New-TestUser
    }

    It "Resets a user's password by sending a reset email to the user" {
        $Response = PSc8y\Reset-UserPassword -Id $User.id -WhatIf 2>&1 | Out-String
        $LASTEXITCODE | Should -Be 0

        $Body = Get-JSONFromResponse $Response

        $Body.sendPasswordResetEmail | Should -BeExactly $true
        $Body.password | Should -BeNullOrEmpty
    }

    It "Resets a user's password by setting a manual password" {
        $pass = New-RandomPassword
        $Response = PSc8y\Reset-UserPassword -Id $User.id -NewPassword $pass -WhatIf 6>&1 | Out-String
        $LASTEXITCODE | Should -Be 0

        $Body = Get-JSONFromResponse $Response

        $Body.password | Should -BeExactly $pass
        $Body.sendPasswordResetEmail | Should -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-User -Id $User.id
    }
}
