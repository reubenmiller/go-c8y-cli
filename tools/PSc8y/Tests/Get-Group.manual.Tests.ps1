. $PSScriptRoot/imports.ps1

Describe -Name "Get-Group" {
    Context "Existing groups" {
        $Group1 = New-TestGroup -Name "tempGroup1"
        $Group2 = New-TestGroup -Name "tempGroup1"

        It "Get a group (using pipeline)" {
            $Response = $Group1, $Group2 | PSc8y\Get-Group

            $LASTEXITCODE | Should -Be 0
            $Response | Should -HaveCount 2
            $Response[0].id | Should -BeExactly $Group1.id
            $Response[1].id | Should -BeExactly $Group2.id
        }

        $null = Remove-Group -Id $Group1.id
        $null = Remove-Group -Id $Group2.id
    }
}
