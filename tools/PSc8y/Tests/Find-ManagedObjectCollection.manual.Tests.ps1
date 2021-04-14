. $PSScriptRoot/imports.ps1

Describe -Name "Find-ManagedObjectCollection" {
    BeforeEach {
        $Device = New-TestDevice -Name "roomUpperFloor_"

    }

    It "Find all devices with their names starting with 'roomUpperFloor_'" {
        $Response = PSc8y\Find-ManagedObjectCollection -Query "name eq 'roomUpperFloor_*'"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Find all devices by piping query text" {
        $output = "name eq '*a*'" | PSc8y\Find-ManagedObjectCollection -OrderBy "_id desc" -OnlyDevices -Dry 2>&1
        $LASTEXITCODE | Should -Be 0
        $requests = $output | ConvertFrom-Json
        $requests | Should -HaveCount 1

        $requests[0] | Should -MatchObject @{
            path = "/inventory/managedObjects";
            query = "query=`$filter=has(c8y_IsDevice) and name eq '*a*' `$orderby=_id desc"
        } -Property path, query
    }

    It "Find matching devices by piping an object" {
        $output = @{ "c8y_DeviceQueryString" = "name eq '*a*'"} | PSc8y\Find-ManagedObjectCollection -OrderBy "_id desc" -OnlyDevices -Dry 2>&1
        $LASTEXITCODE | Should -Be 0
        $requests = $output | ConvertFrom-Json
        $requests | Should -HaveCount 1

        $requests[0] | Should -MatchObject @{
            path = "/inventory/managedObjects";
            query = "query=`$filter=has(c8y_IsDevice) and name eq '*a*' `$orderby=_id desc"
        } -Property path, query
    }

    It "Finds inventory managed objects using a template query" {
        $output = @{ "c8y_DeviceQueryString" = "name eq '*a*'"} | PSc8y\Find-ManagedObjectCollection -OrderBy "_id desc" -QueryTemplate "not(%s)" -Dry 2>&1
        $LASTEXITCODE | Should -Be 0
        $requests = $output | ConvertFrom-Json
        $requests | Should -HaveCount 1

        $requests[0] | Should -MatchObject @{
            path = "/inventory/managedObjects";
            query = "query=`$filter=not(name eq '*a*') `$orderby=_id desc"
        } -Property path, query
    }

    AfterEach {
        Remove-ManagedObject -Id $Device.id

    }
}

