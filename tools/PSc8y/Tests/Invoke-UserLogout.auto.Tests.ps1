. $PSScriptRoot/imports.ps1

Describe -Name "Invoke-UserLogout" {
    BeforeEach {

    }

    It "Log out the current user" {
        $Response = PSc8y\Invoke-UserLogout
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

