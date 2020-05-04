. $PSScriptRoot/imports.ps1

Describe -Name "Update-User" {
    BeforeEach {
        $User = PSc8y\New-TestUser

    }

    It "Update a user" {
        $Response = PSc8y\Update-User -Id $User.id -FirstName "Simon"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-User -Id $User.id

    }
}

