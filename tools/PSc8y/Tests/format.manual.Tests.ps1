. $PSScriptRoot/imports.ps1

Describe -Name "c8y format global parameter" {
    It "returns just the id" {
        $output = c8y applications get --id cockpit --select id --csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -MatchExactly "^\d+$"
    }

    It "returns just the name using wildcard" {
        $output = c8y applications get --id cockpit --select "nam*" --csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -BeExactly "cockpit"
    }

    It "returns id and name" {
        $output = c8y applications get --id cockpit --select "id,name" --csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -MatchExactly "^\d+,cockpit$"
    }

    It "includes empty values for non-existant values" {
        $output = c8y applications get --id cockpit --select "id,name,doesnotexist" --csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -MatchExactly "^\d+,cockpit,$"
    }

    It "includes multiple lines for a list of inputs" {
        $output = c8y applications list --pageSize 2 --select "id,name" --csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
        $output | Should -HaveCount 2
        $output | Should -MatchExactly "^\d+,\w+$"
    }

    It "returns output which can be read via csv (without headers)" {
        $output = c8y applications get --id cockpit --select "id,name,doesnotexist" --csv
        $LASTEXITCODE | Should -Be 0
        $table = $output | ConvertFrom-Csv -Header id, name, unknown
        $table.id | Should -MatchExactly "^\d+$"
        $table.name | Should -MatchExactly "^cockpit$"
        $table.unknown | Should -BeNullOrEmpty
    }

    It "returns just the id using wildcards" {
        $output = c8y applications get --id cockpit --select "id*" --csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -MatchExactly "^\d+$"
    }

    It "returns json lines" {
        $output = c8y applications get --id cockpit --select "id*"
        $LASTEXITCODE | Should -Be 0
        $json = $output | ConvertFrom-Json
        $json."id*" | Should -MatchExactly "^\d+$"
    }

    It "returns json lines with multiple properties" {
        $output = c8y applications get --id cockpit --select "id,name"
        $LASTEXITCODE | Should -Be 0
        $json = $output | ConvertFrom-Json
        $json."id" | Should -MatchExactly "^\d+$"
        $json."name" | Should -MatchExactly "^\w+$"
    }

    It "returns json lines using custom properties names" {
        $output = c8y applications get --id cockpit --select "appId:id,appName:name"
        $LASTEXITCODE | Should -Be 0
        $json = $output | ConvertFrom-Json
        $json."appId" | Should -MatchExactly "^\d+$"
        $json."appName" | Should -MatchExactly "^\w+$"
    }

    It "returns csv ignoring the aliases when no header options is provided" {
        $output = c8y applications get --id cockpit --select "appId:id,appName:name" --csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 1
        $output | Should -MatchExactly "^\d+,\w+$"
    }

    It "filters and selects a subset of properties" {
        $output = c8y applications list --pageSize 100 --filter "name like cockpi*" --select id,name
        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
        $output | Should -HaveCount 1
        $json = $output | ConvertFrom-Json
        $json.id | Should -MatchExactly "^\d+$"
        $json.name | Should -BeExactly "cockpit"
    }

    It -Skip -Tag @("TODO") "devices that do not match the filter are ignored" {
        # Need to create devices to support this test scenario
        $output = c8y devices list | c8y devices get --filter "name like device*" --select id,name,self --workers 5
        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
        $output | Should -HaveCount 3
    }

    It "select properties and csv output" {
        $output = c8y applications list --select '{app Name:id,id:id}'
        $LASTEXITCODE | Should -Be 0
        $json = $output | ConvertFrom-Json
        $json."app Name" | Should -Not -BeNullOrEmpty
        $json.id | Should -Not -BeNullOrEmpty
    }

    AfterEach {
    }
}
