. $PSScriptRoot/imports.ps1

Describe -Name "Reset-UserPassword" {
    BeforeEach {
        $User = PSc8y\New-TestUser
    }

    It "Resets a user's password by sending a reset email to the user" {
        PSc8y\Reset-UserPassword -Id $User.id -WhatIf -InformationVariable Request
        $LASTEXITCODE | Should -Be 0

        $Body = Get-JSONFromResponse ($Request | Out-String)

        $Body.sendPasswordResetEmail | Should -BeExactly $true
        $Body.password | Should -BeNullOrEmpty
    }

    It "Resets a user's password by setting a manual password" {
        $pass = New-RandomPassword
        PSc8y\Reset-UserPassword -Id $User.id -NewPassword $pass -WhatIf -InformationVariable Request
        $LASTEXITCODE | Should -Be 0

        $Body = Get-JSONFromResponse ($Request | Out-String)

        $Body.password | Should -BeExactly $pass
        $Body.sendPasswordResetEmail | Should -BeExactly $false
    }


    AfterEach {
        PSc8y\Remove-User -Id $User.id
    }
}
