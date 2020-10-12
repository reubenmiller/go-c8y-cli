. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Group" {
    Context "Existing groups" {
        BeforeAll {
            $Prefix = New-RandomString -Prefix "tempGroup1_"
            $Group1 = New-TestGroup -Name $Prefix
            $Group2 = New-TestGroup -Name $Prefix
        }

        It "Delete a multiple user groups (using pipeline)" {
            $Response = PSc8y\Get-GroupCollection -PageSize 2000 `
                | Where-Object { $_.name -like "${Prefix}*" } `
                | Remove-Group

            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty

            # wait for deletion
            Start-Sleep -Seconds 5

            [array] $GroupsAfterDeletion = PSc8y\Get-GroupCollection -PageSize 2000 `
                | Where-Object { $_.name -like "${Prefix}*" }

            $GroupsAfterDeletion.Count | Should -BeExactly 0
        }

        AfterAll {
            $null = Remove-Group -Id $Group1.id -ErrorAction SilentlyContinue 2>&1
            $null = Remove-Group -Id $Group2.id -ErrorAction SilentlyContinue 2>&1
        }
    }
}
