. $PSScriptRoot/../imports.ps1

Describe -Name "c8y select global parameter" {
    BeforeAll {
        $ids = New-Object System.Collections.ArrayList
    }


    It "returns output which can be read via csv (without headers)" {
        $output = c8y applications get --id cockpit --select "id,name,doesnotexist" --output csv
        $LASTEXITCODE | Should -Be 0
        $table = $output | ConvertFrom-Csv -Header id, name, unknown
        $table.id | Should -MatchExactly "^\d+$"
        $table.name | Should -MatchExactly "^cockpit$"
        $table.unknown | Should -BeNullOrEmpty
    }

    Context "Large json values" {
        It "handles large json values" {
            $Template = New-TemporaryFile
            Set-Content -Path $Template -Value  @"
local item() = {
    "id": 1,
    "first_name": "Doloritas",
    "last_name": "Clow",
    "email": "dclow0@berkeley.edu",
    "favourite_food": "popcorn",
    "ip_address": "249.218.130.49"
};

{
    "data": [item() for i in std.range(1, 100000)]
}
"@
            $id = c8y inventory create --template $Template --select id --output csv
            [void] $ids.Add($id)
            $Start = Get-Date
            $output = c8y inventory get --id $id
            $LASTEXITCODE | Should -Be 0
            $output | Should -Not -BeNullOrEmpty
            $DurationSec = ((Get-Date) - $Start).TotalSeconds
            $DurationSec | Should -BeLessOrEqual 30
        }

        It "sorts output keys using natural sort" {
            $Template = New-TemporaryFile
            Set-Content -Path $Template -Value  @"
local item(i) = {
    "index": i,
};

{
    "data": [item(i) for i in std.range(0, 20)]
}
"@
            $id = c8y inventory create --template $Template --select id --output csv
            [void] $ids.Add($id)
            $output = c8y inventory get --id $id --select "data.**" --output csvheader
            $LASTEXITCODE | Should -Be 0
            $output[0] | Should -MatchExactly "^data\.0\.index,data\.1\.index,data\.2\.index"
            $output[1] | Should -BeExactly (@(0..20) -join ",")
        }
    }

    AfterAll {
        if ($ids.Count -gt 0) {
            $ids | c8y inventory delete --force
        }
    }
}
