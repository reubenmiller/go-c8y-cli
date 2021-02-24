. $PSScriptRoot/imports.ps1

Describe -Name "Reset-UserPassword" {
    BeforeEach {
        $User = PSc8y\New-TestUser

    }

    It "Resets a user's password by sending a reset email to the user" {
        $Response = PSc8y\Reset-UserPassword -Id $User.id -WhatIf 2>&1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Resets a user's password by generating a new password" {
        $Response = PSc8y\Reset-UserPassword -Id $User.id -NewPassword (New-RandomPassword)
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-User -Id $User.id

    }
}

