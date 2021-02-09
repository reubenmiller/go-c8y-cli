. $PSScriptRoot/imports.ps1

Describe -Name "c8y pipes" {
    BeforeAll {
        Set-Alias -Name c8yb -Value (Get-ClientBinary)

        Function ConvertTo-JsonPipe {
            [cmdletbinding()]
            Param(
                [Parameter(
                    ValueFromPipeline = $true,
                    Position = 0
                )]
                [object[]] $InputObject
            )
            Process {
                $InputObject | ForEach-Object { ConvertTo-Json $_ -Depth 100 -Compress }
            }
        }
    }
    BeforeEach {

    }

    Context "Piping to single commands" {
        It "Pipe by id a simple getter" {
            $output = @("1", "2") | c8yb events get --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET /event/events/1" -Total 1
            $output | Should -ContainRequest "GET /event/events/2" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }

        It "Pipe by id update command" {
            $output = @("1", "2") | c8yb events update --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "PUT" -Total 2
            $output | Should -ContainRequest "PUT /event/events/1" -Total 1
            $output | Should -ContainRequest "PUT /event/events/2" -Total 1
        }

        It "Pipe by id delete command" {
            $output = @("1", "2") | c8yb events delete --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "DELETE" -Total 2
            $output | Should -ContainRequest "DELETE /event/events/1" -Total 1
            $output | Should -ContainRequest "DELETE /event/events/2" -Total 1
        }

        It "Pipe by id create command" {
            $output = @("10", "11") | c8yb events create --template "{type: 'c8y_Event', text: 'custom info ' + input.index}" --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "POST" -Total 2
            $output | Should -ContainRequest "POST /event/events" -Total 2
        }
    }
    
    Context "Piping to collection commands" {

    
        It "Pipe by id to query parameters" {
            $output = @("1", "2") | c8yb events list --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET /event/events?source=1" -Total 1
            $output | Should -ContainRequest "GET /event/events?source=2" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }

        It "Empty pipe. Empty values should not cause a lookup, however they should also not stop the iteration" {
            $output = @("", "") | c8yb events list --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET" -Total 2
            $output | Should -ContainRequest "GET /event/events?source" -Total 0
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }

        It "Pipe by id object to query parameters" {
            $output = @{id=1}, @{id=2} | ConvertTo-JsonPipe | c8yb events list --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET /event/events?source=1" -Total 1
            $output | Should -ContainRequest "GET /event/events?source=2" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }

        It "Pipe by json object using deviceId rather than id to query parameters" {
            $output = @{id=3; deviceId=1}, @{id=4; deviceId=2} | ConvertTo-JsonPipe | c8yb events list --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET" -Total 2
            $output | Should -ContainRequest "GET /event/events?source=1" -Total 1
            $output | Should -ContainRequest "GET /event/events?source=2" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }

        It "Pipe by json object using source.id rather than id to query parameters" {
            $output = @{id=3; source=@{id=1}}, @{id=4; source=@{id=2}} | ConvertTo-JsonPipe | c8yb events list --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET" -Total 2
            $output | Should -ContainRequest "GET /event/events?source=1" -Total 1
            $output | Should -ContainRequest "GET /event/events?source=2" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }

        It "Pipe by name which do not match to query parameters ignoring names that does not exist" {
            $output = @("pipeNameDoesNotExist1", "pipeNameDoesNotExist2", "pipeNameDoesNotExist3") | c8yb events list --dry 2>&1
            $LASTEXITCODE | Should -Be 104
            $output | Should -ContainRequest "GET" -Total 3
            $output | Should -ContainRequest "GET /inventory/managedObjects" -Total 3
            $output | Should -ContainRequest "GET /event/events" -Total 0 -Because "Unresolved names should not trigger queries"
        }

        It "Pipe by name which do not match to query parameters aborts after specified number of errors" {
            $output = @("pipeNameDoesNotExist1", "pipeNameDoesNotExist2", "pipeNameDoesNotExist3", "pipeNameDoesNotExist4") | c8yb events list --dry --abortOnErrors 1 2>&1
            $LASTEXITCODE | Should -Be 103
            $output | Should -ContainRequest "GET" -Minimum 1 -Maximum 2 -Because "Abort is not instantaneous, so lookups can be sent out"
            $output | Should -ContainRequest "GET /inventory/managedObjects" -Minimum 1 -Maximum 2
        }

        It "Pipe by id and name to query parameters. Invalid reference by names should be skipped or should throw an error?" {
            $output = @("1", "pipeNameDoesNotExist2") | c8yb events list --dry 2>&1
            $LASTEXITCODE | Should -Not -Be 0
            $output | Should -ContainRequest "GET" -Total 2
            $output | Should -ContainRequest "GET /event/events" -Total 1
            $output | Should -ContainRequest "GET /event/events?source=1" -Total 1 -Because "Unresolved names should not trigger queries"
            $output | Should -ContainRequest "GET /inventory/managedObjects" -Total 1 -Because "only reference by name lookups use inventory api"
        }

        It "Get results without piped variable" {
            $output = c8yb events list --dry 2>&1
            $LASTEXITCODE | Should -Be 0
            $output | Should -ContainRequest "GET" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 1
            $output | Should -ContainRequest "GET /event/events/source" -Total 0
        }
    }

    AfterEach {
    }
}
