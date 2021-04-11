. $PSScriptRoot/imports.ps1

Describe -Name "powershell pipes" {
    BeforeAll {
        $ids = New-Object System.Collections.ArrayList
        $deviceIds = 1..2 | c8y devices create --select id --output csv
    }

    Context "Get commands" {
        It "Pipe by id a simple getter" {
            $output = ,$deviceIds | Get-ManagedObject -AsPSObject:$false -Debug 2>&1
            $LASTEXITCODE | Should -Be 0
            $output -match "command: c8y" | Should -HaveCount 1
            $output -match "adding job: 2" | Should -HaveCount 1
            $output | Should -ContainRequest "GET /inventory/managedObjects/$($deviceIds[0])" -Total 1
            $output | Should -ContainRequest "GET /inventory/managedObjects/$($deviceIds[1])" -Total 1
            $output | Should -ContainRequest "GET /inventory/managedObjects" -Total 2
        }

        It "Should handle piping objects directly with integer ids" {
            $output = Get-InventoryRoleCollection -PageSize 1 | Get-InventoryRole
            $output | Should -HaveCount 1
        }
    }

    Context "Update commands" {
        It "Pipe by id a update managed object" {
            $output = ,$deviceIds | Update-Device -Data "myvalue=1" -Dry -DryFormat json 2>&1
            $LASTEXITCODE | Should -Be 0
            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2

            $requests[0].method | Should -Be "PUT"
            $requests[0].path | Should -BeExactly "/inventory/managedObjects/$($deviceIds[0])"
            $requests[1].body | Should -MatchObject @{
                myvalue = 1
            }

            $requests[1].method | Should -Be "PUT"
            $requests[1].path | Should -BeExactly "/inventory/managedObjects/$($deviceIds[1])"
            $requests[1].body | Should -MatchObject @{
                myvalue = 1
            }
        }

        It "Pipe by id a update managed object using hashtable as body" {
            $output = ,$deviceIds | Update-Device -Data @{myvalue = 1} -Dry -DryFormat json 2>&1
            $LASTEXITCODE | Should -Be 0
            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2

            $requests[0].method | Should -Be "PUT"
            $requests[0].path | Should -BeExactly "/inventory/managedObjects/$($deviceIds[0])"
            $requests[1].body | Should -MatchObject @{
                myvalue = 1
            }

            $requests[1].method | Should -Be "PUT"
            $requests[1].path | Should -BeExactly "/inventory/managedObjects/$($deviceIds[1])"
            $requests[1].body | Should -MatchObject @{
                myvalue = 1
            }
        }

        InModuleScope -ModuleName PSc8y {
            It "Confirmation handles multiple items" {
                $itemIds = 1..2 | c8y devices create --select id --output csv
                $items = $itemIds | Get-ManagedObject -AsPSObject
                $items.id | c8y devices delete
                $message = Format-ConfirmationMessage -Name "Get-ExampleName" -InputObject $items
                $message | Should -Match $itemIds[0]
                $message | Should -Match $itemIds[1]
            }
        }
    }

    Context "Direct piping" {
        It "Should pipe directly between cmdlets with interger id types" {
            $output = Get-InventoryRoleCollection -PageSize 1 | Get-InventoryRole
            $output | Should -HaveCount 1
        }

        It "Should pipe directly between cmdlets" {
            $output = Get-DeviceCollection -PageSize 4 | Get-ManagedObject
            $output | Should -HaveCount 4
        }
    }

    Context "CSV" {
        It "Should pipe directly between cmdlets" {
            $output = Get-ApplicationCollection -PageSize 5 -Output csv -Select "id,name" | ConvertFrom-CSV -Header id,name
            $output | Should -HaveCount 5
        }
    }

    Context "Device creation" {
        It "accepts devices names from the pipeline" {
            $output = ,@("device01", "device02") | New-Device -Dry -DryFormat json 2>&1
            $LASTEXITCODE | Should -Be 0

            $requests = $output | ConvertFrom-Json
            $requests | Should -HaveCount 2
            $PartialRequest = $requests | Select-Object path, method, body
            $PartialRequest[0] | Should -MatchObject @{ method = "POST"; path = "/inventory/managedObjects" } -Property method, path
            $PartialRequest[0].body | Should -MatchObject @{c8y_IsDevice=@{}; name="device01"}
            
            $PartialRequest[1] | Should -MatchObject @{ method = "POST"; path = "/inventory/managedObjects" } -Property method, path
            $PartialRequest[1].body | Should -MatchObject @{c8y_IsDevice=@{}; name="device02"}
        }
    }

    Context "Colors" {
        It "can pipe colored output into other functions" {
            $output = Get-ManagedObjectCollection -Color -PageSize 1 | Get-ManagedObject
            $LASTEXITCODE | Should -BeExactly 0
            $output | Should -Not -BeNullOrEmpty
        }
    }

    Context "streaming" {
        It "can stream include results to a downstream command in json format" {
            $env:C8Y_SETTINGS_INCLUDEALL_PAGESIZE = 10
            $env:C8Y_SETTINGS_INCLUDEALL_DELAYMS = 1000
            $messages = $( $output = devices -IncludeAll -AsPSObject:$false -TotalPages 3 | batch | Get-Device -Verbose -Delay 0 -Workers 5 -Dry -DryFormat json -WithError ) 2>&1
            $LASTEXITCODE | Should -BeExactly 0
            $messages -match "command: c8y" | Should -HaveCount 1 -Because "all gets should be executed by one c8y call"
            $output.Count | Should -BeGreaterThan 1
            $output.Count | Should -BeLessOrEqual 30
        }
 
        It -Tag "TODO" "streams each page size to the pipeline when it is received and not when all the results are done" {
            $env:C8Y_SETTINGS_INCLUDEALL_PAGESIZE = 10
            $env:C8Y_SETTINGS_INCLUDEALL_DELAYMS = 1000


        }
        $env:C8Y_SETTINGS_INCLUDEALL_DELAYMS = 1000
    }

    Context "Filtering using where-object" {
        It "Filters the output using where-object" {
            $output = Get-RoleCollection -PageSize 100 | Where-Object Name -like "*ALARM*"
            $LASTEXITCODE | Should -BeExactly 0
            $output.Count | Should -BeGreaterThan 1
        }
    }

    Context "View" {
        It "Uses the overall view" {
            $output = devices -WithTotalPages -PageSize 1
            $LASTEXITCODE | Should -BeExactly 0
            $output.psobject.TypeNames[0] | Should -match "collection"
        }
    }

    AfterAll {
        $ids | Remove-ManagedObject
        $deviceIds | Remove-ManagedObject
    }
}
