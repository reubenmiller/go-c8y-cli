. $PSScriptRoot/imports.ps1

Describe -Name "Get-UserMembershipCollection" {
    BeforeEach {
        $User = PSc8y\Get-CurrentUser

    }

    It "Get a list of groups that a user belongs to" {
        $Response = PSc8y\Get-UserMembershipCollection -Id $User.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

