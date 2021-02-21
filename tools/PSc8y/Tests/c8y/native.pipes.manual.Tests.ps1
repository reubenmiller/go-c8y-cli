. $PSScriptRoot/../imports.ps1

Describe -Name "c8y pipes" {
    BeforeAll {
        $ids = New-Object System.Collections.ArrayList
    }

    Context "Piping to single commands" {
        It "Pipe by id a simple getter" {
            $output = @("1", "2") | c8y events get --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET /event/events/1" -Total 1
            $output | Should -ContainRequest "GET /event/events/2" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }

        It "Pipe by id update command" {
            $output = @("1", "2") | c8y events update --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "PUT" -Total 2
            $output | Should -ContainRequest "PUT /event/events/1" -Total 1
            $output | Should -ContainRequest "PUT /event/events/2" -Total 1
        }

        It "Pipe by id delete command" {
            $output = @("1", "2") | c8y events delete --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "DELETE" -Total 2
            $output | Should -ContainRequest "DELETE /event/events/1" -Total 1
            $output | Should -ContainRequest "DELETE /event/events/2" -Total 1
        }

        It "Pipe by id create command" {
            $output = @("10", "11") | c8y events create --template "{type: 'c8y_Event', text: 'custom info ' + input.index}" --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "POST" -Total 2
            $output | Should -ContainRequest "POST /event/events" -Total 2
            $output -like '*"text": "custom info 1"*' | should -HaveCount 1
            $output -like '*"text": "custom info 2"*' | should -HaveCount 1
        }
    }
    
    Context "Piping to collection commands" {

    
        It "Pipe by id to query parameters" {
            $output = @("1", "2") | c8y events list --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET /event/events?source=1" -Total 1
            $output | Should -ContainRequest "GET /event/events?source=2" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }

        It "Empty pipe. Empty values should not cause a lookup, however they should also not stop the iteration" {
            $output = @("", "") | c8y events list --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET" -Total 2
            $output | Should -ContainRequest "GET /event/events?source" -Total 0
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }

        It "Pipe by id object to query parameters" {
            $output = @{id=1}, @{id=2} `
            | Invoke-ClientIterator -AsJSON `
            | c8y events list --dry 2>&1

            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET /event/events?source=1" -Total 1
            $output | Should -ContainRequest "GET /event/events?source=2" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }

        It "Pipe by json object using deviceId rather than id to query parameters" {
            $output = @{id=3; deviceId=1}, @{id=4; deviceId=2} `
            | Invoke-ClientIterator -AsJSON `
            | c8y events list --dry 2>&1

            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET" -Total 2
            $output | Should -ContainRequest "GET /event/events?source=1" -Total 1
            $output | Should -ContainRequest "GET /event/events?source=2" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }

        It "Pipe by json object using source.id rather than id to query parameters" {
            $output = @{id=3; source=@{id=1}}, @{id=4; source=@{id=2}} `
            | Invoke-ClientIterator -AsJSON `
            | c8y events list --dry 2>&1

            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET" -Total 2
            $output | Should -ContainRequest "GET /event/events?source=1" -Total 1
            $output | Should -ContainRequest "GET /event/events?source=2" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }

        It "Pipe by name which do not match to query parameters ignoring names that does not exist" {
            $output = @("pipeNameDoesNotExist1", "pipeNameDoesNotExist2", "pipeNameDoesNotExist3") | c8y events list --dry 2>&1
            $LASTEXITCODE | Should -Be 104
            $output | Should -ContainRequest "GET" -Total 3
            $output | Should -ContainRequest "GET /inventory/managedObjects" -Total 3
            $output | Should -ContainRequest "GET /event/events" -Total 0 -Because "Unresolved names should not trigger queries"
        }

        It "Pipe by name which do not match to query parameters aborts after specified number of errors" {
            $output = @("pipeNameDoesNotExist1", "pipeNameDoesNotExist2", "pipeNameDoesNotExist3", "pipeNameDoesNotExist4") | c8y events list --dry --abortOnErrors 1 2>&1
            $LASTEXITCODE | Should -Be 103
            $output | Should -ContainRequest "GET" -Minimum 1 -Maximum 2 -Because "Abort is not instantaneous, so lookups can be sent out"
            $output | Should -ContainRequest "GET /inventory/managedObjects" -Minimum 1 -Maximum 2
        }

        It "Pipe by id and name to query parameters. Invalid reference by names should be skipped or should throw an error?" {
            $output = @("1", "pipeNameDoesNotExist2") | c8y events list --dry 2>&1
            $LASTEXITCODE | Should -Not -Be 0
            $output | Should -ContainRequest "GET" -Total 2
            $output | Should -ContainRequest "GET /event/events" -Total 1
            $output | Should -ContainRequest "GET /event/events?source=1" -Total 1 -Because "Unresolved names should not trigger queries"
            $output | Should -ContainRequest "GET /inventory/managedObjects" -Total 1 -Because "only reference by name lookups use inventory api"
        }

        It "Get results without piped variable" {
            $output = c8y events list --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 1
            $output | Should -ContainRequest "GET /event/events/source" -Total 0
        }
    }

    Context "Pipe to optional query parameters" {
        It "Pipe an ids to a query parameter" {
            $output = @("1", "2") | c8y events deleteCollection --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "DELETE /event/events?source=1" -Total 1
            $output | Should -ContainRequest "DELETE /event/events?source=2" -Total 1
            $output | Should -ContainRequest "DELETE /event/events" -Total 2
        }

        It "Pipe an ids to a query parameter" {
            $output = c8y events deleteCollection --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "DELETE /event/events?source" -Total 0
            $output | Should -ContainRequest "DELETE /event/events" -Total 1
        }
    }

    Context "Pipe to optional body" {
        It -Skip -Tag @("Deprecated:UsingBatch") "Pipe an ids to a body parameter" {
            $output = @("name1", "name2") | c8y inventory create --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "POST /inventory/managedObjects" -Total 2
        }

        It "No pipe input to a body parameter" {
            $output = c8y inventory create --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "POST /inventory/managedObjects" -Total 1
        }
    }

    Context "pipeline - devices" {
        It "Provides piped strings to template" {
            $output = "11", "12" | c8y devices create --template "{ jobIndex: input.index, jobValue: input.value }" --dry 2>&1
            $LASTEXITCODE | Should -BeExactly 0

            $output | SHould -ContainRequest "POST /inventory/managedObjects" -Total 2
            $Bodies = $output | Get-RequestBodyCollection | Sort-Object jobIndex
            $Bodies | Should -HaveCount 2
            $Bodies[0] | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=1; jobValue="11"; name="11"}
            $Bodies[1] | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=2; jobValue="12"; name="12"}

            $Bodies[0].name | Should -BeOfType [string]
            $Bodies[0].jobIndex | Should -BeOfType [long]
            $Bodies[0].jobValue | Should -BeOfType [string]
        }

        It "Provides piped json to template" {
            $output = @{name="myDevice01"}, @{name="myDevice02"} `
            | Invoke-ClientIterator -AsJSON `
            | c8y devices create --template "{ jobIndex: input.index, name: input.value.name }" --dry 2>&1
            $LASTEXITCODE | Should -BeExactly 0

            $output | Should -ContainRequest "POST /inventory/managedObjects" -Total 2
            $Bodies = $output | Get-RequestBodyCollection | Sort-Object jobIndex

            $Bodies | Should -HaveCount 2
            $Bodies[0] | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=1; name="myDevice01"}
            $Bodies[1] | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=2; name="myDevice02"}

            $Bodies[0].name | Should -BeOfType [string]
            $Bodies[0].jobIndex | Should -BeOfType [long]
        }

        It "pipes objects from powershell to the c8y binary" {
            $output = 1..2 `
            | Invoke-ClientIterator "device" `
            | c8y devices create --template "{ jobIndex: input.index, name: input.value.name }" --dry 2>&1
            $LASTEXITCODE | Should -BeExactly 0

            $output | SHould -ContainRequest "POST /inventory/managedObjects" -Total 2
            $Bodies = $output | Get-RequestBodyCollection | Sort-Object jobIndex

            $Bodies | Should -HaveCount 2
            $Bodies[0] | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=1; name="device0001"}
            $Bodies[1] | Should -MatchObject @{c8y_IsDevice=@{}; jobIndex=2; name="device0002"}

            $Bodies[0].name | Should -BeOfType [string]
            $Bodies[0].jobIndex | Should -BeOfType [long]
        }

        It "supports templates referencing input values: ids => get => update" {
            $device1 = New-TestDevice -Template "{type: 'customType1'}"
            $device2 = New-TestDevice -Template "{type: 'customType2'}"
            $null = $ids.Add($device1.id)
            $null = $ids.Add($device2.id)

            # pipe output 
            $output = $device1.id, $device2.id `
            | c8y devices get `
            | c8y devices update --template "{ type: input.value.type + 'Suffix', index: input.index }" --dry 2>&1
            $LASTEXITCODE | Should -BeExactly 0

            $output | Should -ContainRequest "PUT /inventory/managedObjects" -Total 2
            $Bodies = $output | Get-RequestBodyCollection | Sort-Object index

            $Bodies | Should -HaveCount 2

            $Bodies[0] | Should -MatchObject @{type="customType1Suffix"; index=1}
            $Bodies[0].type | Should -BeOfType [string]
            $Bodies[0].index | Should -BeOfType [long]
            $Bodies[0] | Should -MatchObject @{type="customType2Suffix"; index=2}
            $Bodies[0].type | Should -BeOfType [string]
            $Bodies[0].index | Should -BeOfType [long]
        }
    }

    AfterAll {
        $ids | Remove-ManagedObject
    }
}
