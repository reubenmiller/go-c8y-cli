. $PSScriptRoot/../imports.ps1

Describe -Name "c8y pipes" {
    BeforeAll {
        $DeviceList = @(1..10) | ForEach-Object {
            New-TestAgent
        }
        $ids = New-Object System.Collections.ArrayList
    }

    Context "Job limits" {
        It "stops early due to job limit being exceeded" {
            $output = @("1", "2", "3") | c8y events get --maxJobs 1 --dry --verbose 2>&1
            $LASTEXITCODE | Should -Be 105
            $output | Should -ContainRequest "GET /event/events/1" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 1
        }

        It "stops early due to job limit being exceeded using env variable" {
            $env:C8Y_SETTINGS_DEFAULTS_MAXJOBS = "2"
            $output = @("1", "2", "3") | c8y events get --dry --verbose 2>&1
            $env:C8Y_SETTINGS_DEFAULTS_MAXJOBS = ""
            $LASTEXITCODE | Should -Be 105
            $output | Should -ContainRequest "GET /event/events/1" -Total 1
            $output | Should -ContainRequest "GET /event/events/2" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 2
        }

        It "aborts on job errors" {
            $output = @("NonExistantName1", "NonExistantName2", "NonExistantName3") | c8y events list --abortOnErrors 1 --dry --verbose 2>&1
            $LASTEXITCODE | Should -Be 103
            $output | Should -ContainRequest "GET /inventory/managedObjects" -Minimum 1 -Maximum 2
            $output | Should -ContainRequest "GET /event/events" -Total 0
        }

        It "aborts on job errors piping to non-existant values" {
            # Piping values to an id should not result in lookups!
            $output = @("NonExistantName1", "NonExistantName2", "NonExistantName3") | c8y events list --abortOnErrors 1 --dry --verbose 2>&1
            $LASTEXITCODE | Should -Be 103
            $output | Should -ContainRequest "GET" -Total 1
            $output | Should -ContainRequest "GET /inventory/managedObjects" -Total 1
            $output | Should -ContainRequest "GET /event/events" -Total 0
            ($output -match "aborted batch as error count has been exceeded") | Should -HaveCount 1
        }
    }

    Context "Update-Device" {

        It "Updates device managed object using workers" {   
            $output = $DeviceList.id | c8y devices update --template "{ counter: input.index }" --workers 5
            $LASTEXITCODE | Should -BeExactly 0
            $output | Should -ContainInCollection $DeviceList
            $output | Should -HaveCount $DeviceList.Count
            $json = $output | ConvertFrom-Json -Depth 100
            $json.counter | Should -ContainInCollection @(1..$DeviceList.Count) -Because "Counter should increment over pipeline"
        }
    }

    Context "New-Measurement" {
        It "Creates a fixed measurement for each input device" {
            $c8yargs = @(
                "--time", "0s",
                "--type", "c8y_TestMeasurement",
                "--data", "c8y_TestMeasurement={temperature:{value:1.2345,unit:'°C'}}",
                "--delay", "1",
                "--workers", "1"
            )
            $output = $DeviceList.id | c8y measurements create $c8yargs
            $LASTEXITCODE | Should -BeExactly 0
            $json = $output | ConvertFrom-Json -Depth 100
            ($json.time | Sort-Object -Unique).Count | Should -BeGreaterOrEqual ($DeviceList.Count - 1) -Because "second request is queued as quick as possible"
            $json.source.id | Should -ContainInCollection $DeviceList.id
        }

        It "Creates a fixed measurement for each input device using multiple workers" {
            $c8yargs = @(
                "--time", "0s",
                "--type", "c8y_TestMeasurement",
                "--data", "c8y_TestMeasurement={temperature:{value:1.2345,unit:'°C'}}",
                "--delay", "500",
                "--workers", "2"
            )
            $output = $DeviceList.id | c8y measurements create $c8yargs
            $LASTEXITCODE | Should -BeExactly 0
            $json = $output | ConvertFrom-Json -Depth 100
            ($json.time | Sort-Object -Unique).Count | Should -BeGreaterThan 1
            $json.source.id | Should -ContainInCollection $DeviceList.id
        }

        It "Creates a measurements from a template with workers" {
            $c8yargs = @(
                "--delay", "500",
                "--workers", "2",
                "--template", "{time: time.nowNano, type: 'type_' + input.index ,c8y_TestMeasurement:{temperature:{value:rand.float,unit:'°C'}}}"
            )
            $output = $DeviceList.id | c8y measurements create $c8yargs
            $LASTEXITCODE | Should -BeExactly 0
            $json = $output | ConvertFrom-Json -Depth 100
            $json.type | Sort-Object -Unique | Should -HaveCount $DeviceList.Count
            $json.source.id | Should -ContainInCollection $DeviceList.id
            ($json.time | Sort-Object -Unique).Count | Should -BeGreaterThan 1
        }
    }

    Context "events" {
        It "Creates events from a template" {
            $c8yargs = @(
                "--delay", "500",
                "--workers", "2",
                "--template", "{time: time.nowNano, type: 'type_' + input.index, text: std.format('custom event %03d', input.index) }"
            )
            $output = $DeviceList.id | c8y events create $c8yargs
            $LASTEXITCODE | Should -BeExactly 0
            $json = $output | ConvertFrom-Json -Depth 100
            $json.type | Sort-Object -Unique | Should -HaveCount $DeviceList.Count
            $json.type -match "^type_([0-9]|10)$" | Should -HaveCount $DeviceList.Count
            $json.text -match "^custom event (00[0-9]|010)$" | Should -HaveCount $DeviceList.Count
            $json.source.id | Should -ContainInCollection $DeviceList.id
            ($json.time | Sort-Object -Unique).Count | Should -BeGreaterThan 1
        }

        It "supports the event workflow create -> update -> delete using multiple workers" {
            $c8yargs = @(
                "--delay", "500",
                "--workers", "2",
                "--template", "{time: time.nowNano, type: 'type_' + input.index, text: std.format('custom event %03d', input.index) }"
            )
            # Create
            $events = $DeviceList `
            | Select-Object -First 1 `
            | Invoke-ClientIterator -Repeat 5 `
            | c8y events create $c8yargs `
            | ConvertFrom-Json -Depth 100
            $LASTEXITCODE | Should -BeExactly 0
            $events | Should -HaveCount 5

            # Update
            $output = $events `
            | Invoke-ClientIterator `
            | c8y events update --template "{ text: 'event ' + input.index  }" --workers 5 `
            | ConvertFrom-Json -Depth 100
            $LASTEXITCODE | Should -BeExactly 0
            $output | Should -ContainInCollection $events
            $output.text -match "^event [1-5]$" | Should -HaveCount $events.Count

            # Delete
            $output = $events `
            | Invoke-ClientIterator `
            | c8y events delete --workers 5 --verbose 2>&1
            $LASTEXITCODE | Should -BeExactly 0
            $output | should -ContainRequest "DELETE /event/events/" -Total $events.Count
            $events.id | ForEach-Object {
                $output | should -ContainRequest "DELETE /event/events/$_" -Total 1
            }

            # Delete collection (iterator over device id)
            $output = $DeviceList `
            | Invoke-ClientIterator `
            | c8y events deleteCollection --workers 5 --verbose 2>&1
            $LASTEXITCODE | Should -BeExactly 0
            $output | should -ContainRequest "DELETE /event/events?source=" -Total $DeviceList.Count
            $DeviceList.id | ForEach-Object {
                $output | should -ContainRequest "DELETE /event/events?source=$_" -Total 1
            }
        }
    }

    Context "alarms" {
        It "supports the alarm workflow: create -> update -> deleteCollection using multiple workers" {
            $alarmType = New-RandomString -Prefix "ci_alarm_workers"
            $c8yargs = @(
                "--delay", "500",
                "--workers", "2",
                # -Type c8y_TestAlarm -Time "-0s" -Text "Test alarm" -Severity MAJOR
                "--template", "{time: time.nowNano, type: '$alarmType' + input.index, severity: 'MAJOR', text: std.format('custom alarm %03d', input.index) }"
            )
            # Create
            $alarms = $DeviceList `
            | Select-Object -First 1 -Skip 1 `
            | Invoke-ClientIterator -Repeat 5 `
            | c8y alarms create $c8yargs `
            | ConvertFrom-Json -Depth 100
            $LASTEXITCODE | Should -BeExactly 0
            $alarms | Should -HaveCount 5

            # Update
            $output = $alarms `
            | Invoke-ClientIterator `
            | c8y alarms update --template "{ status: 'ACKNOWLEDGED', text: 'alarm ' + input.index  }" --workers 5 `
            | ConvertFrom-Json -Depth 100
            $LASTEXITCODE | Should -BeExactly 0
            $output | Should -ContainInCollection $alarms
            $output.text -match "^alarm [1-5]$" | Should -HaveCount $alarms.Count
            $output.status -match "ACKNOWLEDGED" | Should -HaveCount $alarms.Count

            # Delete alarms for each device
            # Note: delete by id is not support by Cumulocity and will return a 405 status code
            $output = $DeviceList `
            | Invoke-ClientIterator `
            | c8y alarms deleteCollection --workers 5 --verbose 2>&1
            $LASTEXITCODE | Should -BeExactly 0
            $output | should -ContainRequest "DELETE /alarm/alarms" -Total $DeviceList.Count
            $DeviceList.id | ForEach-Object {
                $output | should -ContainRequest "DELETE /alarm/alarms?source=$_" -Total 1
            }
        }
    }

    Context "operations" {
        It "supports the operation workflow: create -> update -> delete using multiple workers" {
            $c8yargs = @(
                "--delay", "500",
                "--workers", "2",
                "--template", "{c8y_Restart: {}, description: std.format('custom operation %03d', input.index) }"
            )
            # Create
            $operations = $DeviceList `
            | Select-Object -First 1 -Skip 1 `
            | Invoke-ClientIterator -Repeat 5 `
            | c8y operations create $c8yargs `
            | ConvertFrom-Json -Depth 100
            $LASTEXITCODE | Should -BeExactly 0
            $operations | Should -HaveCount 5

            # Update
            $output = $operations `
            | Invoke-ClientIterator `
            | c8y operations update --template "{ status: 'EXECUTING', description: 'updated operation ' + input.index  }" --workers 5 `
            | ConvertFrom-Json -Depth 100
            $LASTEXITCODE | Should -BeExactly 0
            $output | Should -ContainInCollection $operations
            $output.description -match "^updated operation [1-5]$" | Should -HaveCount $operations.Count
            $output.status -match "EXECUTING" | Should -HaveCount $operations.Count

            # Delete operations for each device
            # Note: delete by id is not support by Cumulocity and will return a 405 status code
            $output = $DeviceList `
            | Invoke-ClientIterator `
            | c8y operations deleteCollection --workers 5 --verbose 2>&1
            $LASTEXITCODE | Should -BeExactly 0
            $output | should -ContainRequest "DELETE /devicecontrol/operations" -Total $DeviceList.Count
            $DeviceList.id | ForEach-Object {
                $output | should -ContainRequest "DELETE /devicecontrol/operations?deviceId=$_" -Total 1
            }
        }
    }

    Context "Devices" {
        It "supports the device workflow: create -> update -> delete using multiple workers" {
            $c8yargs = @(
                "--delay", "500",
                "--workers", "2",
                "--template", "{name: std.format('testdevice-%03d', input.index) }"
            )
            # Create
            $devices = @(1..5) `
            | ForEach-Object { c8y devices create $c8yargs } `
            | ConvertFrom-Json -Depth 100
            $null = $ids.AddRange($devices.id)
            $LASTEXITCODE | Should -BeExactly 0
            $devices | Should -HaveCount 5
            $devices.name -match "testdevice-00[1-5]" | Should -HaveCount $devices.Count

            # Add devices to a group
            $group = New-TestDeviceGroup
            $output = $devices `
            | Invoke-ClientIterator `
            | c8y devicegroups assignDevice --group $group.id --workers 5 --raw -o json `
            | ConvertFrom-Json -Depth 100
            $LASTEXITCODE | Should -BeExactly 0
            $output | Should -HaveCount $devices.Count
            $output.managedObject | Should -ContainInCollection $devices
            $devices.id | ForEach-Object {
                $output.self -like "*/inventory/managedObjects/$($group.id)/childAssets/$_" | Should -HaveCount 1
            }

            # Get child assets
            $childAssets = $group `
            | Invoke-ClientIterator `
            | c8y devicegroups listAssets `
            | ConvertFrom-Json -Depth 100
            $LASTEXITCODE | Should -BeExactly 0
            $childAssets | Should -HaveCount $devices.Count

            # Unassign assets from group
            $output = $childAssets `
            | Invoke-ClientIterator `
            | c8y devicegroups unassignDevice --group $group.id --workers 5 --verbose 2>&1
            $LASTEXITCODE | Should -BeExactly 0
            $output | Should -ContainRequest "DELETE" -Total $childAssets.Count
            $childAssets.id | ForEach-Object {
                $output | Should -ContainRequest "DELETE /inventory/managedObjects/$($group.id)/childAssets/$_" -Total 1
            }

            # $devices = $DeviceList `
            # | Select-Object -First 1 -Skip 1 `
            # | Invoke-ClientIterator -Repeat 5 `
            # | c8y devices create $c8yargs `
            # | ConvertFrom-Json -Depth 100
            # $LASTEXITCODE | Should -BeExactly 0
            # $devices | Should -HaveCount 5
            

            # Update
            $output = $devices `
            | Invoke-ClientIterator `
            | c8y devices update --template "{ type: 'updated device ' + input.index }" --workers 5 `
            | ConvertFrom-Json -Depth 100
            $LASTEXITCODE | Should -BeExactly 0
            $output | Should -ContainInCollection $devices
            $output.type -match "^updated device [1-5]$" | Should -HaveCount $devices.Count

            # Delete
            $output = $devices `
            | Invoke-ClientIterator `
            | c8y devices delete --workers 5 --verbose 2>&1
            $LASTEXITCODE | Should -BeExactly 0
            $output | should -ContainRequest "DELETE /inventory/managedObjects/" -Total $devices.Count
            $devices.id | ForEach-Object {
                $output | should -ContainRequest "DELETE /inventory/managedObjects/$_" -Total 1
            }
        }
    }

    Context "managed objects" {
        It "supports creating many managed objects using multiple workers" {
            $c8yargs = @(
                "--delay", "50",
                "--workers", "5",
                # "--template", "{type: 'type_' + input.index }"
                "--template", "{type: std.format('type_%03d', input.index) }"
            )
            # Create
            $devices = @(1..5) | Invoke-ClientIterator "testdevice_" `
            | c8y devices create $c8yargs --workers 5 --maxJobs 10 `
            | ConvertFrom-Json -Depth 100
            $null = $ids.AddRange($devices.id)
            $LASTEXITCODE | Should -BeExactly 0
            $devices | Should -HaveCount 5
            $devices.name -match "testdevice_000[1-5]" | Should -HaveCount $devices.Count
            $devices.type -match "type_00[1-5]" | Should -HaveCount $devices.Count
            
            # Remove devices
            # $devices `
            # | Invoke-ClientIterator `
            # | c8y devices delete --workers 5
            # $LASTEXITCODE | Should -BeExactly 0

        }
    }

    AfterAll {
        $DeviceList | Remove-Device
        $ids | Remove-ManagedObject -ErrorAction SilentlyContinue
    }
}
