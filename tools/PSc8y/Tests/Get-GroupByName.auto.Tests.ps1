. $PSScriptRoot/imports.ps1

Describe -Name "Get-GroupByName" {
    BeforeEach {
        $Group = New-TestGroup

    }

    It "Get user group by its name" {
        $Response = PSc8y\Get-GroupByName -Name $Group.name
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Group -Id $Group.id

    }
}

