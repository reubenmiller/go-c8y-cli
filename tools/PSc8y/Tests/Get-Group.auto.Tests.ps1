. $PSScriptRoot/imports.ps1

Describe -Name "Get-Group" {
    BeforeEach {
        $Group = New-TestGroup

    }

    It "Get a user group" {
        $Response = PSc8y\Get-Group -Id $Group.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Group -Id $Group.id

    }
}

