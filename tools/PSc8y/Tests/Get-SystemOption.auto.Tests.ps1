. $PSScriptRoot/imports.ps1

Describe -Name "Get-SystemOption" {
    BeforeEach {

    }

    It "Get system option value" {
        $Response = PSc8y\Get-SystemOption -Category "system" -Key "version"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

