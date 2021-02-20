. $PSScriptRoot/../imports.ps1

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
        $type = New-RandomString -Prefix "selectType"
        1..2 | c8y devices create --data "type=$type,text=value"
        3 | c8y devices create --data "type=$type"
        $output = c8y devices list --type $type --select "name,id,text,type" --csv
        c8y devices list --type $type | c8y devices delete
        $LASTEXITCODE | Should -Be 0
        $output = $output | Sort-Object
        $output | Should -HaveCount 3
        $output[0] | Should -MatchExactly "^1,\d+,value,$type$"
        $output[1] | Should -MatchExactly "^2,\d+,value,$type$"
        $output[2] | Should -MatchExactly "^3,\d+,,$type$"
    }

    It "includes empty values for non-existant values in the last field" {
        $type = New-RandomString -Prefix "selectType"
        1 | c8y devices create --data "type=$type"
        $output = c8y devices list --type $type --select "name,id,type,nonexistant" --csv
        c8y devices list --type $type | c8y devices delete
        $LASTEXITCODE | Should -Be 0
        $output | Should -MatchExactly "^1,\d+,$type,$"
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
        $json.id | Should -MatchExactly "^\d+$"
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

    Context "flat selection" {
        It "Does not produce duplicate json keys" {
            $output = c8y applications list --select "id,*" --pageSize 1 --compact=false
            ($output -match '"id":') | Should -HaveCount 1
        }

        It "does not match partial key names if no wildcards are used" {
            $output = c8y applications list --select "nam" --pageSize 1
            $LASTEXITCODE | Should -BeExactly 0
            $output | Should -Not -Match '"name":'
        }

        It "should return no results when the select property does not match" {
            $output = c8y applications list --select "asdfasdf" --pageSize 1 --compact=false
            $LASTEXITCODE | Should -BeExactly 0
            ($output | Out-String).Trim() | Should -BeExactly "{}"
        }

        It "select a nested object by name only" {
            $output = c8y applications list --select "id,owner" --pageSize 1 --type MICROSERVICE --compact=false
            $LASTEXITCODE | Should -BeExactly 0
            $json = $output | ConvertFrom-Json -Depth 100
            $json | Should -HaveCount 1
            $json.id | Should -Not -BeNullOrEmpty
            $json.owner | Should -Not -BeNullOrEmpty
            $json.owner.self | Should -Not -BeNullOrEmpty
            $json.owner.tenant | Should -Not -BeNullOrEmpty
            $json.owner.tenant.id | Should -Not -BeNullOrEmpty
        }

        It "matches all nested properties when using globstar suffix" {
            $output = c8y applications list --select "owner.tena***" --pageSize 1 --type MICROSERVICE --compact=false
            $LASTEXITCODE | Should -BeExactly 0
            $json = $output | ConvertFrom-Json -Depth 100
            $json | Should -HaveCount 1
            $json.owner.tenant.id | Should -Not -BeNullOrEmpty
        }

        It "matches all nested properties when using globstar suffix" {
            $output = c8y applications list --select "owner.tenant" --pageSize 1 --type MICROSERVICE --compact=false
            $LASTEXITCODE | Should -BeExactly 0
            $json = $output | ConvertFrom-Json -Depth 100
            $json | Should -HaveCount 1
            $json.owner.tenant | Should -Not -BeNullOrEmpty
            $json.owner.tenant.id | Should -Not -BeNullOrEmpty
        }

        It "matches all propteries which end with the nested property structure using globstar prefix" {
            $output = c8y applications list --select "**tenant.id" --pageSize 1 --type MICROSERVICE --compact=false
            $LASTEXITCODE | Should -BeExactly 0
            $json = $output | ConvertFrom-Json
            $json.owner.tenant.id | Should -Not -BeNullOrEmpty
        }

        It "selects only select level properties" {
            $output = c8y applications list --select "owner.*" --pageSize 1 --compact=false
            $LASTEXITCODE | Should -BeExactly 0
            $json = $output | ConvertFrom-Json -Depth 100
            $json.owner | Should -Not -BeNullOrEmpty
            $json.owner.self | Should -Not -BeNullOrEmpty
            $json.id | Should -BeNullOrEmpty
        }

        It "selects only select level properties" {
            $output = c8y applications list --select "self" --pageSize 1 --compact=false
            $LASTEXITCODE | Should -BeExactly 0
            $output -match 'self' | Should -HaveCount 1
        }

        It "selects arrays" {
            $output = c8y applications list --type MICROSERVICE --pageSize 1 --select "id,name,roles" --compact=false
            $LASTEXITCODE | Should -BeExactly 0
            $output -match 'roles' | Should -HaveCount 1
            $json = $output | ConvertFrom-Json
            $json.id | Should -Not -BeNullOrEmpty
            $json.name | Should -Not -BeNullOrEmpty
            $json.roles | Should -Not -BeNullOrEmpty
            $json.roles.Count | Should -BeGreaterThan 0
        }
    }

    AfterEach {
    }
}
