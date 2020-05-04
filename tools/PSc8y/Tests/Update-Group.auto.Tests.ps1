. $PSScriptRoot/imports.ps1

Describe -Name "Update-Group" {
    BeforeEach {
        $Group = New-TestGroup

    }

    It "Update a user group" {
        $Response = PSc8y\Update-Group -Id $Group -Name "customGroup2"
        $LASTEXITCODE | Should -Be 0
    }

    It "Update a user group (using pipeline)" {
        $Response = PSc8y\Get-GroupByName -Name $Group.name | Update-Group -Name "customGroup2"
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-Group -Id $Group.id

    }
}

