. $PSScriptRoot/../imports.ps1

Describe -Name "c8y smartgroups" {
    BeforeEach {
        $prefix = New-RandomString
    }

    It "creates a smart group" {
        $output = c8y smartgroups create --name $prefix --query "name eq '$prefix'" --dry --dryFormat json
        $LASTEXITCODE | Should -Be 0
        $requests = $output | ConvertFrom-Json
        $requests | Should -HaveCount 1
        $requests.body | Should -MatchObject @{
            name = $prefix
            c8y_IsDynamicGroup = @{}
            type = 'c8y_DynamicGroup'
            c8y_DeviceQueryString = "name eq '$prefix'"
        }
    }

    It "creates a smart group via pipeline" {
        $output = "name eq '$prefix'" | c8y smartgroups create --name $prefix --dry --dryFormat json
        $LASTEXITCODE | Should -Be 0
        $requests = $output | ConvertFrom-Json
        $requests | Should -HaveCount 1
        $requests.body | Should -MatchObject @{
            name = $prefix
            c8y_IsDynamicGroup = @{}
            type = 'c8y_DynamicGroup'
            c8y_DeviceQueryString = "name eq '$prefix'"
        }
    }

    It "creates a smart group via pipeline but overriding query" {
        $output = "name eq '$prefix'" | c8y smartgroups create --query "has(c8y_IsDevice)" --name $prefix --dry --dryFormat json
        $LASTEXITCODE | Should -Be 0
        $requests = $output | ConvertFrom-Json
        $requests | Should -HaveCount 1
        $requests.body | Should -MatchObject @{
            name = $prefix
            c8y_IsDynamicGroup = @{}
            type = 'c8y_DynamicGroup'
            c8y_DeviceQueryString = "has(c8y_IsDevice)"
        }
    }

    AfterEach {
    }
}
