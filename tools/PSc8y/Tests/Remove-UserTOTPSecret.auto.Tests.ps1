. $PSScriptRoot/imports.ps1

Describe -Name "Remove-UserTOTPSecret" {
    BeforeEach {
        $User = PSc8y\New-TestUser

    }

    It -Skip "Revoke a user's TOTP (TFA) secret" {
        $Response = PSc8y\Remove-UserTFA -Id $User.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

