. $PSScriptRoot/imports.ps1

Describe -Name "powershell pipes" {
    BeforeAll {
        $ids = New-Object System.Collections.ArrayList
        $deviceIds = 1..2 | c8y devices create --select id --csv
    }

    Context "Get commands" {
        It "Pipe by id a simple getter" {
            $output = $deviceIds | pipe | Get-ManagedObject -AsJSON -Verbose 2>&1
            $LASTEXITCODE | Should -Be 0
            $output -match "Loaded session:" | Should -HaveCount 1
            $output -match "adding job: 2" | Should -HaveCount 1
            $output | Should -ContainRequest "GET /inventory/managedObjects/$($deviceIds[0])" -Total 1
            $output | Should -ContainRequest "GET /inventory/managedObjects/$($deviceIds[1])" -Total 1
            $output | Should -ContainRequest "GET /inventory/managedObjects" -Total 2
        }

        It "Should handle piping objects directly with integer ids" {
            $output = Get-CurrentUserInventoryRoleCollection -PageSize 1 | pipe | Get-CurrentUserInventoryRole
            $output | Should -HaveCount 1
        }
    }

    Context "Update commands" {
        It "Pipe by id a update managed object" {
            $output = $deviceIds | pipe | Update-Device -Data "myvalue=1" -AsJSON -Verbose 2>&1
            $LASTEXITCODE | Should -Be 0
            $output -match "Loaded session:" | Should -HaveCount 1
            $output -match "adding job: 2" | Should -HaveCount 1
            $output | Should -ContainRequest "PUT /inventory/managedObjects/$($deviceIds[0])" -Total 1
            $output | Should -ContainRequest "PUT /inventory/managedObjects/$($deviceIds[1])" -Total 1
            $output | Should -ContainRequest "PUT /inventory/managedObjects" -Total 2
        }

        It "Pipe by id a update managed object using hashtable as body" {
            $output = $deviceIds | pipe | Update-Device -Data @{myvalue = 1} -AsJSON -Verbose 2>&1
            $LASTEXITCODE | Should -Be 0
            $output -match "Loaded session:" | Should -HaveCount 1
            $output -match "adding job: 2" | Should -HaveCount 1
            $output | Should -ContainRequest "PUT /inventory/managedObjects/$($deviceIds[0])" -Total 1
            $output | Should -ContainRequest "PUT /inventory/managedObjects/$($deviceIds[1])" -Total 1
            $output | Should -ContainRequest "PUT /inventory/managedObjects" -Total 2
        }

        InModuleScope -ModuleName PSc8y {
            It "Confirmation handles multiple items" {
                $itemIds = 1..2 | c8y devices create --select id --csv
                $items = $itemIds | Get-ManagedObject | pipe
                $items | c8y devices delete
                $message = Format-ConfirmationMessage -Name "Get-ExampleName" -InputObject $items
                $message | Should -Match $itemIds[0]
                $message | Should -Match $itemIds[1]
            }
        }
    }

    Context "Direct piping" {
        It "Should pipe directly between cmdlets" {
            $output = Get-CurrentUserInventoryRoleCollection -PageSize 1 | Get-CurrentUserInventoryRole
            $output | Should -HaveCount 1
        }
    }

    Context "CSV" {
        It "Should pipe directly between cmdlets" {
            $output = Get-ApplicationCollection -PageSize 5 -AsCSV -Select "id,name" | ConvertFrom-CSV -Header id,name
            $output | Should -HaveCount 5
        }
    }

    Context "Device creation" {
        It "accepts devices names from the pipeline" {
            $output = "device01", "device02" | pipe | New-Device -WhatIf 2>&1
            $LASTEXITCODE | Should -Be 0
            $output -match "Loaded session:" | Should -HaveCount 1
            $output -match "adding job: 2" | Should -HaveCount 1
            $output | Should -ContainRequest "POST /inventory/managedObjects" -Total 2

            $Bodies = $output | Get-RequestBodyCollection | Sort-Object name
            $Bodies | Should -HaveCount 2
            $Bodies[0] | Should -MatchObject @{c8y_IsDevice=@{}; name="device01"}
            $Bodies[1] | Should -MatchObject @{c8y_IsDevice=@{}; name="device02"}
        }
    }

    Context "Colors" {
        It "can pipe colored output into other functions" {
            $output = Get-ManagedObjectCollection -Color -PageSize 1 | pipe | Get-ManagedObject
            $LASTEXITCODE | Should -BeExactly 0
            $output | Should -Not -BeNullOrEmpty
        }
    }

    AfterAll {
        $ids | Remove-ManagedObject
        $deviceIds | Remove-ManagedObject
    }
}
