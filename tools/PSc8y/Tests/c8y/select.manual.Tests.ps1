. $PSScriptRoot/../imports.ps1

Describe -Name "c8y select global parameter" {
    BeforeAll {
        $ids = New-Object System.Collections.ArrayList
    }
    It "returns just the id" {
        $output = c8y applications get --id cockpit --select id --output csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -MatchExactly "^\d+$"
    }

    It "returns just the name using wildcard" {
        $output = c8y applications get --id cockpit --select "nam*" --output csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -BeExactly "cockpit"
    }

    It "returns id and name" {
        $output = c8y applications get --id cockpit --select "id,name" --output csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -MatchExactly "^\d+,cockpit$"
    }

    It "includes empty values for non-existant values" {
        $type = New-RandomString -Prefix "selectType"
        1..2 | c8y devices create --data "type=$type,text=value"
        3 | c8y devices create --data "type=$type"
        $output = c8y devices list --type $type --select "name,id,text,type" --output csv
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
        $output = c8y devices list --type $type --select "name,id,type,nonexistant" --output csv
        c8y devices list --type $type | c8y devices delete
        $LASTEXITCODE | Should -Be 0
        $output | Should -MatchExactly "^1,\d+,$type,$"
    }

    It "includes multiple lines for a list of inputs" {
        $output = c8y applications list --pageSize 2 --select "id,name" --output csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
        $output | Should -HaveCount 2
        $output | Should -MatchExactly "^\d+,\w+$"
    }

    It "returns output which can be read via csv (without headers)" {
        $output = c8y applications get --id cockpit --select "id,name,doesnotexist" --output csv
        $LASTEXITCODE | Should -Be 0
        $table = $output | ConvertFrom-Csv -Header id, name, unknown
        $table.id | Should -MatchExactly "^\d+$"
        $table.name | Should -MatchExactly "^cockpit$"
        $table.unknown | Should -BeNullOrEmpty
    }

    It "returns just the id using wildcards" {
        $output = c8y applications get --id cockpit --select "id*" --output csv
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

    It "includes empty objects in the response" {
        $type = New-RandomString -Prefix "withEmptyFragments"
        $device = "test1234" | c8y devices create --type $type
        $output = c8y devices list --type $type --select "**" --output "json"
        $exitcode = $LASTEXITCODE
        $device | c8y devices delete

        $exitcode | Should -Be 0
        $json = $output | ConvertFrom-Json
        $json.c8y_IsDevice | ConvertTo-Json -Compress | Should -BeExactly "{}"
    }


    It "returns csv ignoring the aliases when no header options is provided" {
        $output = c8y applications get --id cockpit --select "appId:id,appName:name" --output csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 1
        $output | Should -MatchExactly "^\d+,\w+$"
    }

    It "returns csv with custom column headers based on aliases" {
        $output = c8y applications get --id cockpit --select "appId:id,appName:name" --output csvheader
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output[0] | Should -MatchExactly "^appId,appName$"
        $output[1] | Should -MatchExactly "^\d+,\w+$"
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

    It "devices that do not match the filter are ignored" {
        $type = New-RandomString -Prefix "selectWithAlias"
        "device01" | c8y devices create --type "$type"
        $output = c8y devices list | c8y devices get --filter "name like asdf*" --select id,name,self --workers 5
        c8y devices list --type $type | c8y devices delete

        $LASTEXITCODE | Should -Be 0
        $output | Should -BeNullOrEmpty
    }

    It "select properties and csv output" {
        $output = c8y applications list --select '{app Name:id,id:id}'
        $LASTEXITCODE | Should -Be 0
        $json = $output | ConvertFrom-Json
        $json."app Name" | Should -Not -BeNullOrEmpty
        $json.id | Should -Not -BeNullOrEmpty
    }

    Context "aliases, data shapping" {
        It "returns json lines using custom properties names" {
            $output = c8y applications get --id cockpit --select "appId:id,appName:name"
            $LASTEXITCODE | Should -Be 0
            $json = $output | ConvertFrom-Json
            $json."appId" | Should -MatchExactly "^\d+$"
            $json."appName" | Should -MatchExactly "^\w+$"
        }
    
        It "returns json lines using custom properties names for nested values" {
            # $output = c8y applications get --id cockpit --select "appId:id,tenantId:owner.**.id"
            $output = c8y applications get --id cockpit --select "appId:id,tenantId:owner.**.id,tenantDetails:owner.**"
            $LASTEXITCODE | Should -Be 0
            $json = $output | ConvertFrom-Json
            $json.appId | Should -MatchExactly "^\d+$"
            $json.tenantId | Should -BeExactly "management"
            $json.tenantDetails.self | Should -Match "^\w+://.+management$"
            $json.tenantDetails.tenant.id | Should -BeExactly "management"
        }

        It "adds nested objects under a propery name when using globstar **" {
            $type = New-RandomString -Prefix "selectWithAlias"
            1 | c8y devices create --type "$type"
            $output = c8y devices list --type $type --select "id:id,links:**.self"
            c8y devices list --type $type | c8y devices delete
            $LASTEXITCODE | Should -Be 0
            $json = $output | ConvertFrom-Json
            $json.id | Should -MatchExactly "^\d+$"
            $json.links | Should -Not -BeNullOrEmpty
            $json.links.deviceParents.self | Should -Not -BeNullOrEmpty
            $json.links.assetParents.self | Should -Not -BeNullOrEmpty
            $json.links.childDevices.self | Should -Not -BeNullOrEmpty
        }

        It "maps nested properties to a new name" {
            $type = New-RandomString -Prefix "selectWithAlias"
            1 | c8y devices create --type "$type" --template "{c8y_Details: {data: {name: 'one'}}}"
            $output1 = c8y devices list --type $type --select "id:id,details:c8y_Details.**"
            $output2 = c8y devices list --type $type --select "id:id,details:c8y_Detail*.**"
            
            c8y devices list --type $type | c8y devices delete
            $LASTEXITCODE | Should -Be 0
            
            $json = $output1 | ConvertFrom-Json
            $json.id | Should -MatchExactly "^\d+$"
            $json.details | Should -Not -BeNullOrEmpty
            $json.details.data.name | Should -Not -BeNullOrEmpty

            $json = $output2 | ConvertFrom-Json
            $json.id | Should -MatchExactly "^\d+$"
            $json.details | Should -Not -BeNullOrEmpty
            $json.details.data.name | Should -Not -BeNullOrEmpty
        }

        It "maps properties to a new property name including the root property" {
            $type = New-RandomString -Prefix "selectWithAlias"
            1 | c8y devices create --type "$type" --template "{c8y_Details: {data: {name: 'one'}}}"
            $output = c8y devices list --type $type --select "id:id,details:**.c8y_Details.**"
            c8y devices list --type $type | c8y devices delete
            $LASTEXITCODE | Should -Be 0
            $json = $output | ConvertFrom-Json
            $json.id | Should -MatchExactly "^\d+$"
            $json.details.c8y_Details | Should -Not -BeNullOrEmpty
            $json.details.c8y_Details.data.name | Should -Not -BeNullOrEmpty
        }

        It "maps nested properties and only literals to a new property name" {
            $type = New-RandomString -Prefix "selectWithAlias"
            1 | c8y devices create --type "$type" --template "{c8y_Details: {name: 'two', data: {name: 'one'}}}"
            $output = c8y devices list --type $type --select "id:id,details:c8y_Details.*"
            c8y devices list --type $type | c8y devices delete
            $LASTEXITCODE | Should -Be 0
            $json = $output | ConvertFrom-Json
            $json.id | Should -MatchExactly "^\d+$"
            $json.details.name | Should -BeExactly "two"
            $json.details.data | Should -BeNullOrEmpty
        }

        It "handles duplicates keys by returning both of the matches" {
            $type = New-RandomString -Prefix "selectWithAlias"
            1 | c8y devices create --type $type --template "{value: 1, Value: 2}"
            $output1 = c8y devices list --type $type --select "value,Value"
            $output2 = c8y devices list --type $type --select "value:Value"
            c8y devices list --type $type | c8y devices delete
            $LASTEXITCODE | Should -Be 0
            
            $json = $output1 | ConvertFrom-Json -AsHashtable
            $json.value | Should -BeExactly 1
            $json.Value | Should -BeExactly 2

            # Note: Duplicate values overwrite eachother when using aliases
            $json = $output2 | ConvertFrom-Json -AsHashtable
            $json.value | Should -BeExactly 1 -Because "alias values are not case sensitive, and value comes after Value when keys are sorted"
        }
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

        It "returns only objects on the selected level with the self property" {
            $type = New-RandomString -Prefix "selectLevel"
            1 | c8y devices create --type "$type"
            $output1 = c8y devices list --type $type --select "*.self"
            $output2 = c8y devices list --type $type --select "**.self"
            c8y devices list --type $type | c8y devices delete
            $LASTEXITCODE | Should -Be 0

            # Depth 2 self links
            $json1 = $output1 | ConvertFrom-Json
            $json1.additionParents.self | Should -Not -BeNullOrEmpty
            $json1.assetParents.self | Should -Not -BeNullOrEmpty
            $json1.self | Should -BeNullOrEmpty -Because "no globstar was used, so depth matching is explicit by number of dots"

            # all self links
            $json2 = $output2 | ConvertFrom-Json
            $json2.additionParents.self | Should -Not -BeNullOrEmpty
            $json2.assetParents.self | Should -Not -BeNullOrEmpty
            $json2.self | Should -Not -BeNullOrEmpty  -Because "globstar was used, so matching can occur on any depth"
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

    Context "Object with number like keys" {
        It "Objects with numbers as keys should not be converted to an array" {
            $DashboardJSON = '{"c8y_Dashboard":{"15426326034650895":{"name":"test"}}}'
    
            $Start = Get-Date
            $output = $DashboardJSON | c8y util show -v --select "**"
            $Duration = ((Get-Date) - $Start).TotalSeconds
            $Duration | Should -BeLessThan 10
            $json = $output | ConvertFrom-Json
            $json.c8y_Dashboard."15426326034650895".name | Should -BeExactly "test"
        }
    
        It "Objects with numbers as keys should not be converted to an array" {
            $DashboardJSON = '{"c8y_Dashboard":{"15426326034650895":{"name":"test"}}}'
    
            $Start = Get-Date
            $output = $DashboardJSON | c8y util show -v --select "c8y_Dashboard.*15426326034650895.**"
            $Duration = ((Get-Date) - $Start).TotalSeconds
            $Duration | Should -BeLessThan 10
            $json = $output | ConvertFrom-Json
            $json.c8y_Dashboard."15426326034650895".name | Should -BeExactly "test"
        }
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
