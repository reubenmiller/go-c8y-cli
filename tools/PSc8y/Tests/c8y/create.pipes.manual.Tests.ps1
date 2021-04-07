. $PSScriptRoot/../imports.ps1

Describe -Name "creating data with pipes" {
    BeforeAll {
        $ids = New-Object System.Collections.ArrayList
        $commonArgs = @(
            "--dry",
            "--dryFormat", "json",
            "--withError"
        )
    }

    Context "Applications" {

        
        It "Accepts simple arguments (no pipeline)" {
            $output = c8y applications create $commonArgs --name "mynewapp" --template "{key: self.name + '-key'}" --type MICROSERVICE
            $LASTEXITCODE | Should -Be 0
            $request = $output | ConvertFrom-Json

            $request.path | Should -BeExactly "/application/applications"
            $request.body | Should -MatchObject @{
                key = "mynewapp-key"
                name = "mynewapp"
                type = "MICROSERVICE"
            }
        }

        It "Overrides a piped value with an explicit argument" {
            $output = "name01" | c8y applications create $commonArgs --name "mynewapp" --template "{key: self.name + '-key'}" --type MICROSERVICE
            $LASTEXITCODE | Should -Be 0
            $request = $output | ConvertFrom-Json

            $request.path | Should -BeExactly "/application/applications"
            $request.body | Should -MatchObject @{
                key = "mynewapp-key"
                name = "mynewapp"
                type = "MICROSERVICE"
            }
        }

        It "Accepts json complex value (override using argument)" {
            $pipedInput = ConvertTo-Json -Compress -InputObject @{
                requiredRoles = @(
                    "EXAMPLE_ROLE_1",
                    "EXAMPLE_ROLE_2"
                )
            }
            $output = $pipedInput | c8y applications create $commonArgs --name "mynewapp" --template "input.value + { key: self.name + '-key'}" --type MICROSERVICE
            $LASTEXITCODE | Should -Be 0
            $request = $output | ConvertFrom-Json

            $request.path | Should -BeExactly "/application/applications"
            $request.body | Should -MatchObject @{
                key = "mynewapp-key"
                name = "mynewapp"
                requiredRoles = @(
                    "EXAMPLE_ROLE_1",
                    "EXAMPLE_ROLE_2"
                )
                type = "MICROSERVICE"
            }
        }

        It "Accepts piped json lines from stdin" {
            $pipedInput = @(
                '{"key":"my-app1-key","name":"my-app1","type":"MICROSERVICE"}',
                '{"key":"my-app2-key","name":"my-app2","type":"MICROSERVICE"}',
                '{"key":"my-app3-key","name":"my-app3","type":"MICROSERVICE"}',
                '{"key":"my-app4-key","name":"my-app4","type":"MICROSERVICE"}'
            )
            $output = $pipedInput | c8y applications create $commonArgs --template "input.value"
            $LASTEXITCODE | Should -Be 0
            $requests = $output | ConvertFrom-Json

            $requests | Should -HaveCount 4

            $requests[0].path | Should -BeExactly "/application/applications"
            $requests[0].body | Should -MatchObject @{
                key = "my-app1-key"
                name = "my-app1"
                type = "MICROSERVICE"
            }

            $requests[1].path | Should -BeExactly "/application/applications"
            $requests[1].body | Should -MatchObject @{
                key = "my-app2-key"
                name = "my-app2"
                type = "MICROSERVICE"
            }

            $requests[2].path | Should -BeExactly "/application/applications"
            $requests[2].body | Should -MatchObject @{
                key = "my-app3-key"
                name = "my-app3"
                type = "MICROSERVICE"
            }

            $requests[3].path | Should -BeExactly "/application/applications"
            $requests[3].body | Should -MatchObject @{
                key = "my-app4-key"
                name = "my-app4"
                type = "MICROSERVICE"
            }
        }

        It "Accepts a file containing json lines as an argument" {
            $inputFile = New-TemporaryFile
            $pipedInput = @(
                '{"key":"my-app1-key","name":"my-app1","type":"MICROSERVICE"}',
                '{"key":"my-app2-key","name":"my-app2","type":"MICROSERVICE"}',
                '{"key":"my-app3-key","name":"my-app3","type":"MICROSERVICE"}',
                '{"key":"my-app4-key","name":"my-app4","type":"MICROSERVICE"}'
            )
            $pipedInput | Out-File $inputFile
            
            $output = c8y applications create $commonArgs --name $inputFile --template "input.value"
            $LASTEXITCODE | Should -Be 0
            Remove-Item $inputFile
            $requests = $output | ConvertFrom-Json

            $requests | Should -HaveCount 4

            $requests[0].path | Should -BeExactly "/application/applications"
            $requests[0].body | Should -MatchObject @{
                key = "my-app1-key"
                name = "my-app1"
                type = "MICROSERVICE"
            }

            $requests[1].path | Should -BeExactly "/application/applications"
            $requests[1].body | Should -MatchObject @{
                key = "my-app2-key"
                name = "my-app2"
                type = "MICROSERVICE"
            }

            $requests[2].path | Should -BeExactly "/application/applications"
            $requests[2].body | Should -MatchObject @{
                key = "my-app3-key"
                name = "my-app3"
                type = "MICROSERVICE"
            }

            $requests[3].path | Should -BeExactly "/application/applications"
            $requests[3].body | Should -MatchObject @{
                key = "my-app4-key"
                name = "my-app4"
                type = "MICROSERVICE"
            }
        }

        It "Accepts a file containing names as an argument" {
            $inputFile = New-TemporaryFile
            $pipedInput = @(
                'my-app1',
                'my-app2',
                'my-app3',
                'my-app4'
            )
            $pipedInput | Out-File $inputFile
            
            $output = c8y applications create $commonArgs --name $inputFile --template "{ key: self.name  + '-key', type: 'MICROSERVICE' }"
            $LASTEXITCODE | Should -Be 0
            Remove-Item $inputFile
            $requests = $output | ConvertFrom-Json

            $requests | Should -HaveCount 4

            $requests[0].path | Should -BeExactly "/application/applications"
            $requests[0].body | Should -MatchObject @{
                key = "my-app1-key"
                name = "my-app1"
                type = "MICROSERVICE"
            }

            $requests[1].path | Should -BeExactly "/application/applications"
            $requests[1].body | Should -MatchObject @{
                key = "my-app2-key"
                name = "my-app2"
                type = "MICROSERVICE"
            }

            $requests[2].path | Should -BeExactly "/application/applications"
            $requests[2].body | Should -MatchObject @{
                key = "my-app3-key"
                name = "my-app3"
                type = "MICROSERVICE"
            }

            $requests[3].path | Should -BeExactly "/application/applications"
            $requests[3].body | Should -MatchObject @{
                key = "my-app4-key"
                name = "my-app4"
                type = "MICROSERVICE"
            }
        }

        It "Accepts piped data from other c8y commands" {
            $output = c8y applications list --pageSize 1 | c8y applications get
            $LASTEXITCODE | Should -Be 0
            $request = $output | ConvertFrom-Json
            $request.id | Should -Match "^\d+$"
            $request.name | Should -Match "^.+$"
            $request.type | Should -Match "^.+$"
        }
    }

    Context "Devices" {
        
    }

    Context "Alarms" {
        It "Accepts piped devices when creating alarms" {
            $device01 = c8y devices create --name "pipedDevice01" --select id --output csv
            [void] $ids.Add($device01)
            $output = c8y devices get --id $device01 | c8y alarms create $commonArgs --severity CRITICAL --text "my alarm" --type "myType" 
            $LASTEXITCODE | Should -Be 0
            $request = $output | ConvertFrom-Json
            $request.body | Should -MatchObject @{
                text = "my alarm"
                type = "myType"
                severity = "CRITICAL"
                source = @{
                    id = $device01
                }
            } -ExcludeProperty "time"
            $request.path | Should -BeExactly "/alarm/alarms"
            $request.method | Should -Be "POST"
        }

        It "Can duplicate alarms on the same device (changing type)" {
            $device01 = c8y devices create --name "pipedDevice01" --select id --output csv
            [void] $ids.Add($device01)
            $null = $device01 | c8y alarms create --severity CRITICAL --text "my alarm" --type "myType"
            $LASTEXITCODE | Should -Be 0

            Start-Sleep -Seconds 1

            $output = c8y alarms list --device $device01 --select "status,text,time,type,severity,source.*" | c8y alarms create $commonArgs --template "input.value" --type "myType2" 
            $LASTEXITCODE | Should -Be 0
            $request = $output | ConvertFrom-Json
            $request.body | Should -MatchObject @{
                text = "my alarm"
                type = "myType2"
                status = "ACTIVE"
                severity = "CRITICAL"
                source = @{
                    id = $device01
                }
            } -ExcludeProperty "time"
            $request.path | Should -BeExactly "/alarm/alarms"
            $request.method | Should -Be "POST"
        }

        It "Can Copy an alarm from one source to another" {
            # Create source
            $device01 = c8y devices create --name "pipedDevice01" --select id --output csv
            [void] $ids.Add($device01)
            
            # Create destination
            $device02 = c8y devices create --name "pipedDevice02" --select id --output csv
            [void] $ids.Add($device02)

            # Create alarm on source
            $null = $device01 | c8y alarms create --severity CRITICAL --text "my alarm" --type "myType"
            $LASTEXITCODE | Should -Be 0

            Start-Sleep -Seconds 1

            $output = c8y alarms list --device $device01 --select "status,text,time,type,severity" | c8y alarms create $commonArgs --device $device02 --template "input.value" 
            $LASTEXITCODE | Should -Be 0
            $request = $output | ConvertFrom-Json
            $request.body | Should -MatchObject @{
                text = "my alarm"
                type = "myType"
                status = "ACTIVE"
                severity = "CRITICAL"
                source = @{
                    id = $device02
                }
            } -ExcludeProperty "time"
            $request.path | Should -BeExactly "/alarm/alarms"
            $request.method | Should -Be "POST"
        }
    }

    AfterAll {
        $ids | Remove-ManagedObject
    }
}
