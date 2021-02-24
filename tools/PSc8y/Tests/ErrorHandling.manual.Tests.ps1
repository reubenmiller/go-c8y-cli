. $PSScriptRoot/imports.ps1

Describe -Name "Error handling" {
    BeforeAll {
        # keep list of app ids to delete after tests
        $ids = New-Object System.Collections.ArrayList
    }

    It "Returns a server error on ErrorVariable" {

        $response = Get-ManagedObject -Id 0 -ErrorVariable c8yError
        $response | Should -BeNullOrEmpty
        $LASTEXITCODE | Should -Not -Be 0
        $c8yError | Should -Not -BeNullOrEmpty
        $c8yError.Count | Should -BeGreaterOrEqual 10
        $c8yError | Select-Object -Last 1 | Should -Match "^serverError:.+Not Found"
    }

    It "Redirects errors to response" {
        $response = Get-ManagedObject -Id 0 2>&1
        $LASTEXITCODE | Should -Not -Be 0

        # Cast exception to string
        "$response" | Should -Match "Not found"
        $response.Exception.Message | Should -Not -BeNullOrEmpty
    }

    It "Redirects errors to response and ErrorVariable" {
        $c8yError = $( $response = Get-ManagedObject -Id 0 -Verbose )
        $LASTEXITCODE | Should -Not -Be 0

        # Variable can also
        $c8yError.Count | Should -BeGreaterOrEqual 10
        $c8yError | Select-Object -Last 1 | Should -BeExactly $response
    }

    It "sets the exit code based on the HTTP status code" {
        $response = Get-ManagedObject -Id 0 -ErrorVariable c8yError -ErrorAction SilentlyContinue
        $LASTEXITCODE | Should -BeExactly 4 -Because "Exit code 4 = Status Code 404"
        $response | Should -BeExactly $null

        # Variable can also
        $c8yError.Count | Should -BeGreaterOrEqual 10
        $c8yError[-1] | Should -Match "Not Found"
    }

    It "custom client requests do not pipe response to error variable" {
        $response = Invoke-ClientRequest `
            -Uri "alarm/alarms" `
            -Data @{
                "text" = "my example text"
            } `
            -Method "POST" `
            -ErrorVariable c8yError -ErrorAction SilentlyContinue | ConvertFrom-Json
        $LASTEXITCODE | Should -BeExactly 22 -Because "Exit code 22 = Status Code 422 invalid format"
        $response.error | Should -Match "validationError"

        $c8yError | Should -Not -BeNullOrEmpty
    }

    It "produces verbose output" {
        $VerboseMessages = $( $null = Get-ManagedObjectCollection -Verbose ) 2>&1
        @($VerboseMessages -like "*Sending request*") | Should -HaveCount 1
    }

    It "saves whatif information to a variable" {
        $requestInfo = $( $response = New-ManagedObject -Name "My Name" -WhatIf ) 2>&1

        $response | Should -BeNullOrEmpty
        $requestInfo | Should -Not -BeNullOrEmpty
        $requestInfo -match "What If" | Should -HaveCount 1
        $requestInfo -match "Sending \[POST\] request to" | Should -HaveCount 1
        $requestInfo -match "Headers:" | Should -HaveCount 1
        $requestInfo -match "Body:" | Should -HaveCount 1
    }

    It "redirects whatif information standard output" {
        $requestInfo = New-ManagedObject -Name "My Name" -WhatIf 2>&1

        $requestInfo | Should -Not -BeNullOrEmpty
        $requestInfo -match "What If" | Should -HaveCount 1
        $requestInfo -match "Sending \[POST\] request to" | Should -HaveCount 1
        $requestInfo -match "Headers:" | Should -HaveCount 1
        $requestInfo -match "Body:" | Should -HaveCount 1
    }

    AfterAll {
        # Cleanup all managed objects
        if ($ids.Count -gt 0) {
            $ids | Remove-ManagedObject
        }
    }
}

InModuleScope PSc8y {
    Describe "In module tests" {
        It "Throws an error on invalid arguments" {
            $Parameters = @{
                "InvalidParameter" = "1"
            }
            $response = Invoke-ClientCommand `
                -Noun "inventory" `
                -Verb "get" `
                -Parameters $Parameters `
                -Type "application/vnd.com.nsn.cumulocity.inventory+json" `
                -ItemType "" `
                -ResultProperty "" `
                -ErrorVariable c8yError `
                -Raw:$false
            
            $response | Should -BeNullOrEmpty
            $LASTEXITCODE | Should -Not -Be 0
            $c8yError | Should -HaveCount 2
            $c8yError[-1] | Should -Match '^commandError: unknown flag: --invalidParameter'
        }
    }
}
