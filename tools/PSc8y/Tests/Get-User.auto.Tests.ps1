. $PSScriptRoot/imports.ps1

Describe -Name "Get-User" {
    BeforeEach {
        $User = PSc8y\New-TestUser

    }

    It "Get a user" {
        $Response = PSc8y\Get-User -Id $User.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-User -Id $User.id

    }
}

