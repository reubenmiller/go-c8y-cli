. $PSScriptRoot/../imports.ps1

Describe -Name "c8y pipes" {
    BeforeAll {
        $ids = New-Object System.Collections.ArrayList
    }

    Context "Piping to single commands" {
        It "Pipe by id a simple getter" {
            $output = @("1", "2") | c8y events get --dry --dryFormat json
            $LASTEXITCODE | Should -Be 0
            
            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2
            $requests[0] | Should -MatchObject @{method = "GET"; pathEncoded = "/event/events/1"} -Property method, pathEncoded
            $requests[1] | Should -MatchObject @{method = "GET"; pathEncoded = "/event/events/2"} -Property method, pathEncoded
        }

        It "Pipe by id update command" {
            $output = @("1", "2") | c8y events update --dry --dryFormat json
            $LASTEXITCODE | Should -Be 0

            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2
            $requests[0] | Should -MatchObject @{method = "PUT"; pathEncoded = "/event/events/1"} -Property method, pathEncoded
            $requests[1] | Should -MatchObject @{method = "PUT"; pathEncoded = "/event/events/2"} -Property method, pathEncoded
        }

        It "Pipe by id delete command" {
            $output = @("1", "2") | c8y events delete --dry
            $LASTEXITCODE | Should -Be 0

            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2
            $requests[0] | Should -MatchObject @{method = "DELETE"; pathEncoded = "/event/events/1"} -Property method, pathEncoded
            $requests[1] | Should -MatchObject @{method = "DELETE"; pathEncoded = "/event/events/2"} -Property method, pathEncoded
        }

        It "Pipe by id create command" {
            $output = @("10", "11") | c8y events create --template "{type: 'c8y_Event', text: 'custom info ' + input.index}" --dry --dryFormat json
            $LASTEXITCODE | Should -Be 0
            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2
            
            $requests[0] | Should -MatchObject @{method = "POST"; path = "/event/events"} -Property method, path
            $requests[0].body.text | Should -BeExactly "custom info 1"
            $requests[1] | Should -MatchObject @{method = "POST"; path = "/event/events"} -Property method, path
            $requests[1].body.text | Should -BeExactly "custom info 2"
        }
    }
    
    Context "Piping to collection commands" {

    
        It "Pipe by id to query parameters" {
            $output = @("1", "2") | c8y events list --dry --dryFormat json
            $LASTEXITCODE | Should -Be 0
            $requests = $output | ConvertFrom-Json

            $requests | Should -HaveCount 2
            $partial = $requests | Select-Object path, method

            $requests[0].query | Should -Match "source=1"
            $partial[0] | Should -MatchObject @{method="GET"; path="/event/events"}
            
            $requests[1].query | Should -Match "source=2"
            $partial[1] | Should -MatchObject @{method="GET"; path="/event/events"}
        }

        It "Empty pipe. Empty values should not cause a lookup, however they should also not stop the iteration" {
            $output = @("", "") | c8y events list --dry --dryFormat json
            $LASTEXITCODE | Should -Be 0

            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2
            $requests[0] | Should -MatchObject @{method = "GET"; pathEncoded = "/event/events"} -Property method, pathEncoded
            $requests[1] | Should -MatchObject @{method = "GET"; pathEncoded = "/event/events"} -Property method, pathEncoded
        }

        It "Pipe by id object to query parameters" {
            $output = @{id=1}, @{id=2} `
            | Invoke-ClientIterator -AsJSON `
            | c8y events list --dry --dryFormat json

            $LASTEXITCODE | Should -Be 0
            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2
            
            $requests[0] | Should -MatchObject @{method="GET"; path="/event/events"} -Property method, path
            $requests[0].query | Should -Match "source=1"

            $requests[1] | Should -MatchObject @{method="GET"; path="/event/events"} -Property method, path
            $requests[1].query | Should -Match "source=2"
        }

        It "Pipe by json object using deviceId rather than id to query parameters" {
            $output = @{id=3; deviceId=1}, @{id=4; deviceId=2} `
            | Invoke-ClientIterator -AsJSON `
            | c8y events list --dry --dryFormat json

            $LASTEXITCODE | Should -Be 0
            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2
            
            $requests[0] | Should -MatchObject @{method="GET"; path="/event/events"} -Property method, path
            $requests[0].query | Should -Match "source=1"

            $requests[1] | Should -MatchObject @{method="GET"; path="/event/events"} -Property method, path
            $requests[1].query | Should -Match "source=2"
        }

        It "Pipe by json object using source.id rather than id to query parameters" {
            $output = @{id=3; source=@{id=1}}, @{id=4; source=@{id=2}} `
            | Invoke-ClientIterator -AsJSON `
            | c8y events list --dry --dryFormat json

            $LASTEXITCODE | Should -Be 0
            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2
            
            $requests[0] | Should -MatchObject @{method="GET"; path="/event/events"} -Property method, path
            $requests[0].query | Should -Match "source=1"

            $requests[1] | Should -MatchObject @{method="GET"; path="/event/events"} -Property method, path
            $requests[1].query | Should -Match "source=2"
        }

        It "Pipe by name which do not match to query parameters ignoring names that does not exist" {
            $output = @("pipeNameDoesNotExist1", "pipeNameDoesNotExist2", "pipeNameDoesNotExist3") | c8y events list --dry --dryFormat markdown --abortOnErrors 5 --verbose 2>&1
            $LASTEXITCODE | Should -Be 104

            $output | Should -ContainRequest "GET" -Total 3
            $output | Should -ContainRequest "GET /inventory/managedObjects" -Total 3
            $output | Should -ContainRequest "GET /event/events" -Total 0 -Because "Unresolved names should not trigger queries"
        }

        It "Pipe by name which do not match to query parameters aborts after specified number of errors" {
            $inputItems = @("pipeNameDoesNotExist1", "pipeNameDoesNotExist2", "pipeNameDoesNotExist3", "pipeNameDoesNotExist4")
            
            $output = $inputItems | c8y events list --verbose --dry --dryFormat json --abortOnErrors 1 2>&1
            $LASTEXITCODE | Should -Be 103
            $output | Should -ContainRequest "GET" -Minimum 1 -Maximum 2 -Because "Abort is not instantaneous, so lookups can be sent out"
            $output | Should -ContainRequest "GET /inventory/managedObjects" -Minimum 1 -Maximum 2
        }

        It "Pipe by id and name to query parameters. Invalid reference by names should be skipped or should throw an error?" {
            $output = @("1", "pipeNameDoesNotExist2") | c8y events list --dry --dryFormat markdown --verbose 2>&1
            $LASTEXITCODE | Should -Not -Be 0
            $output | Should -ContainRequest "GET" -Total 2
            $output | Should -ContainRequest "GET /event/events" -Total 1
            $output | Should -ContainRequest "GET /event/events?source=1" -Total 1 -Because "Unresolved names should not trigger queries"
            $output | Should -ContainRequest "GET /inventory/managedObjects" -Total 1 -Because "only reference by name lookups use inventory api"
        }

        It "Get results without piped variable" {
            $output = c8y events list --dry --dryFormat json
            $LASTEXITCODE | Should -Be 0

            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 1
            $requests[0] | Should -MatchObject @{method = "GET"; pathEncoded = "/event/events"} -Property method, pathEncoded
        }
    }

    Context "Pipe to optional query parameters" {
        It "Pipe an ids to a query parameter" {
            $output = @("1", "2") | c8y events deleteCollection --dry --dryFormat json
            $LASTEXITCODE | Should -Be 0
            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2
            $requests[0] | Should -MatchObject @{method = "DELETE"; path = "/event/events"; query = "source=1"} -Property method, path, query
            $requests[1] | Should -MatchObject @{method = "DELETE"; path = "/event/events"; query = "source=2"} -Property method, path, query
        }

        It "Pipe an ids to a query parameter" {
            $output = c8y events deleteCollection --dry --dryFormat json
            $LASTEXITCODE | Should -Be 0
            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 1
            $requests[0] | Should -MatchObject @{method = "DELETE"; pathEncoded = "/event/events" } -Property method, pathEncoded
        }
    }

    Context "Pipe to optional body" {
        It "Pipe values to a body parameter" {
            $output = @("name1", "name2") | c8y inventory create --dry --dryFormat json
            $LASTEXITCODE | Should -Be 0
            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2
            $requests[0] | Should -MatchObject @{method = "POST"; path = "/inventory/managedObjects"; body = @{name = "name1"}} -Property method, path, body
            $requests[1] | Should -MatchObject @{method = "POST"; path = "/inventory/managedObjects"; body = @{name = "name2"}} -Property method, path, body
        }

        It "No pipe input to an optional body parameter" {
            $output = c8y inventory create --dry --dryFormat json
            $LASTEXITCODE | Should -Be 0

            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 1
            $requests[0] | Should -MatchObject @{method = "POST"; path = "/inventory/managedObjects"; body = @{}} -Property method, path, body
        }
    }

    Context "pipeline - devices" {
        It "Provides piped strings to template" {
            $output = "11", "12" | c8y devices create --template "{ jobIndex: input.index, jobValue: input.value }" --dry --dryFormat json
            $LASTEXITCODE | Should -BeExactly 0
            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2

            $requests[0] | Should -MatchObject @{method = "POST"; path = "/inventory/managedObjects"} -Property method, path
            $requests[1] | Should -MatchObject @{method = "POST"; path = "/inventory/managedObjects"} -Property method, path
            $requests[0].body | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=1; jobValue="11"; name="11"}
            $requests[1].body | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=2; jobValue="12"; name="12"}

            $requests[0].body.name | Should -BeOfType [string]
            $requests[0].body.jobIndex | Should -BeOfType [long]
            $requests[0].body.jobValue | Should -BeOfType [string]
        }

        It "Provides piped json to template" {
            $output = @{name="myDevice01"}, @{name="myDevice02"} `
            | Invoke-ClientIterator -AsJSON `
            | c8y devices create --template "{ jobIndex: input.index, name: input.value.name }" --dry --dryFormat json
            $LASTEXITCODE | Should -BeExactly 0

            $requests = $output | ConvertFrom-Json
            
            $requests | Should -HaveCount 2
            $requests[0] | Should -MatchObject @{method = "POST"; path = "/inventory/managedObjects"} -Property method, path
            $requests[1] | Should -MatchObject @{method = "POST"; path = "/inventory/managedObjects"} -Property method, path

            $requests[0].body | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=1; name="myDevice01"}
            $requests[1].body | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=2; name="myDevice02"}

            $requests.body[0].name | Should -BeOfType [string]
            $requests.body[0].jobIndex | Should -BeOfType [long]
        }

        It "pipes objects from powershell to the c8y binary" {
            $output = 1..2 `
            | Invoke-ClientIterator "device" `
            | c8y devices create --template "{ jobIndex: input.index, name: input.value.name }" --dry --dryFormat json
            $LASTEXITCODE | Should -BeExactly 0
            $requests = $output | ConvertFrom-Json

            $requests | Should -HaveCount 2
            $requests[0] | Should -MatchObject @{method = "POST"; path = "/inventory/managedObjects"} -Property method, path
            $requests[1] | Should -MatchObject @{method = "POST"; path = "/inventory/managedObjects"} -Property method, path

            $requests[0].body | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=1; name="device0001"}
            $requests[1].body | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=2; name="device0002"}

            $requests[0].body.name | Should -BeOfType [string]
            $requests[0].body.jobIndex | Should -BeOfType [long]
        }

        It "supports templates referencing input values: ids => get => update" {
            $device1 = New-Device -Name "testDevice01" -Template "{type: 'customType1'}"
            $device2 = New-Device -Name "testDevice02" -Template "{type: 'customType2'}"
            $null = $ids.Add($device1.id)
            $null = $ids.Add($device2.id)

            # pipe output 
            $output = $device1.id, $device2.id `
            | c8y devices get `
            | c8y devices update --template "{ type: input.value.type + 'Suffix', index: input.index }" --dry --dryFormat json
            $LASTEXITCODE | Should -BeExactly 0

            $requests = $output | ConvertFrom-Json

            $requests | Should -HaveCount 2
            $requests[0] | Should -MatchObject @{method = "PUT"; path = "/inventory/managedObjects/$($device1.id)"} -Property method, path
            $requests[0].body | Should -MatchObject @{type="customType1Suffix"; index=1}

            $requests[1] | Should -MatchObject @{method = "PUT"; path = "/inventory/managedObjects/$($device2.id)"} -Property method, path
            $requests[1].body | Should -MatchObject @{type="customType2Suffix"; index=2}

            $requests[0].body.type | Should -BeOfType [string]
            $requests[0].body.index | Should -BeOfType [long]

            $requests[1].body.type | Should -BeOfType [string]
            $requests[1].body.index | Should -BeOfType [long]
        }
    }

    AfterAll {
        $ids | Remove-ManagedObject
    }
}
