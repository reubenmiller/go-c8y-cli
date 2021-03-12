. $PSScriptRoot/../imports.ps1

Describe -Name "c8y format" {

    It "Displays output as csv" {
        $output = c8y applications list --output csvheader --pageSize 5
        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
        $items = $output | ConvertFrom-Csv
        $items | Should -HaveCount 5
    }

    It "Displays output as json" {
        $output = c8y applications list --output json --pageSize 5
        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
        $items = $output | ConvertFrom-Json
        $items | Should -HaveCount 5
    }

    It "Displays output as a table" {
        $output = c8y applications list --output table --pageSize 5
        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
        $output[0] | Should -Match "^id\s+name\s+key\s+"
        $output[1] | Should -Match "^--\s+----\s+---\s+"
        $output[2] | Should -Match "^\d+\s+\S+\s+\S+\s+"
        $csv = $output | ConvertFrom-Csv -Delimiter "`t"
        $csv | Should -HaveCount (1+5)
    }

    It "Displays output as a table with custom columns" {
        $output = c8y applications list --output table --pageSize 5 --select id,name
        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
        $output[0] | Should -Match "^id\s+name\s+$"
        $output[1] | Should -Match "^--\s+----\s+$"
        $output[2] | Should -Match "^\d+\s+\S+\s+$"
        $csv = $output | ConvertFrom-Csv -Delimiter "`t"
        $csv | Should -HaveCount (1+5)
    }

    AfterEach {
    }
}
