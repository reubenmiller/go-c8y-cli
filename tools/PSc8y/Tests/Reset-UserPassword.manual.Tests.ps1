. $PSScriptRoot/imports.ps1

Describe -Name "Reset-UserPassword" {
    BeforeEach {
        $User = PSc8y\New-TestUser
    }

    It "Resets a user's password by sending a reset email to the user" {
        $output = PSc8y\Reset-UserPassword -Id $User.id -WhatIf 2>&1
        $LASTEXITCODE | Should -Be 0

        $Bodies = Get-RequestBodyCollection $output
        $Bodies | Should -MatchObject @{
            sendPasswordResetEmail = $true
        }
    }

    It "Resets a user's password by setting a manual password" {
        $pass = New-RandomPassword
        $output = PSc8y\Reset-UserPassword -Id $User.id -NewPassword $pass -WhatIf 2>&1
        $LASTEXITCODE | Should -Be 0

        $Bodies = Get-RequestBodyCollection $output
        $Bodies | Should -MatchObject @{
            password = $pass
            sendPasswordResetEmail = $false
        }
    }

    AfterEach {
        PSc8y\Remove-User -Id $User.id
    }
}
