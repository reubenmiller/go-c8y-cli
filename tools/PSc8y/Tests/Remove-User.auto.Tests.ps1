. $PSScriptRoot/imports.ps1

Describe -Name "Remove-User" {
    BeforeEach {
        $User = PSc8y\New-TestUser

    }

    It "Delete a user" {
        $Response = PSc8y\Remove-User -Id $User.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

