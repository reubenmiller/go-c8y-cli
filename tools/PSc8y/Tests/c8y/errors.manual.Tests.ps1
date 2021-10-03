. $PSScriptRoot/../imports.ps1

Describe -Name "c8y errors" {
    BeforeAll {
        $ids = New-Object System.Collections.ArrayList
    }

    Context "server errors" {

        It "supports ternary operations on the exit code" {
            $name = New-RandomString

            # check if device exists by name if not create it
            $output = c8y devices get --id $name --silentStatusCodes 404 || c8y devices create --name $name

            $mo = $output | ConvertFrom-Json
            $mo[0].name | Should -BeExactly $name

            # run the command again, this time the name should exist so it should not be created
            $output = c8y devices get --id $name --silentStatusCodes 404 || c8y devices create --name $name
            $mo2 = $output | ConvertFrom-Json
            $mo2.id | Should -BeExactly $mo.id -Because "id should already exist"
        }

        It "deletes multiple ids but dont care if they dont exist" {
            $id = c8y inventory create --select id --output csv

            # check if device exists by name if not create it
            $output = @("0", $id) | c8y inventory delete --silentStatusCodes 404 2>&1
            $LASTEXITCODE | Should -Be 4
            $output | Should -HaveCount 0

            $output = @("0", "0") | c8y inventory delete --silentStatusCodes 404 --silentExit 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -HaveCount 0

            $output = @("0", "0") | c8y inventory delete 2>&1
            $LASTEXITCODE | Should -Be 104
            $output | Should -HaveCount 3 -Because "2 x errors + 1 summary error"
        }
    }

    AfterAll {
        $ids | Remove-ManagedObject
    }
}
