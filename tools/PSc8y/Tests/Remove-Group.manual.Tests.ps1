. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Group" {
    Context "Existing groups" {
        BeforeAll {
            $Group1 = New-TestGroup -Name "tempGroup1"
            $Group2 = New-TestGroup -Name "tempGroup1"
        }

        It "Delete a multiple user groups (using pipeline)" {
            $Response = PSc8y\Get-GroupCollection -PageSize 2000 `
                | Where-Object { $_.name -like "tempGroup1*" } `
                | Remove-Group

            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty

            [array] $GroupsAfterDeletion = PSc8y\Get-GroupCollection -PageSize 2000 `
                | Where-Object { $_.name -like "tempGroup1*" }

            $GroupsAfterDeletion.Count | Should -BeExactly 0
        }

        AfterAll {
            $null = Remove-Group -Id $Group1.id -ErrorAction SilentlyContinue 2>&1
            $null = Remove-Group -Id $Group2.id -ErrorAction SilentlyContinue 2>&1
        }
    }
}
