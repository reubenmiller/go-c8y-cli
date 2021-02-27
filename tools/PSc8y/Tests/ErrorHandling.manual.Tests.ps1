. $PSScriptRoot/imports.ps1

Describe -Name "Error handling" {
    BeforeAll {
        # keep list of app ids to delete after tests
        $ids = New-Object System.Collections.ArrayList
    }

    It "Returns a server error on ErrorVariable" {

        $c8yError = $( $response = Get-ManagedObject -Id 0 -Verbose ) 2>&1
        $response | Should -BeNullOrEmpty
        $LASTEXITCODE | Should -Not -Be 0
        $c8yError | Should -Not -BeNullOrEmpty
        $c8yError.Count | Should -BeGreaterOrEqual 10
        $c8yError | Select-Object -Last 1 | Should -Match "serverError.+404"
    }

    It "Redirects errors to response" {
        $response = Get-ManagedObject -Id 0 2>&1
        $LASTEXITCODE | Should -Not -Be 0

        # Cast exception to string
        "$response" | Should -Match "No managedObject for id"
        $response.Exception.Message | Should -Not -BeNullOrEmpty
    }

    It "sets the exit code based on the HTTP status code" {
        $c8yError = $( $response = Get-ManagedObject -Id 0 ) 2>&1
        $LASTEXITCODE | Should -BeExactly 4 -Because "Exit code 4 = Status Code 404"
        $response | Should -BeExactly $null

        # Variable can also
        $c8yError | Should -HaveCount 1
        $c8yError[-1] | Should -Match "No managedObject for id"
    }

    It "custom client requests do not pipe response to error variable" {
        $response = Invoke-ClientRequest `
            -Uri "alarm/alarms" `
            -Data @{
                "text" = "my example text"
            } `
            -Method "POST" `
            -WithError
        $LASTEXITCODE | Should -BeExactly 22 -Because "Exit code 22 = Status Code 422 invalid format"
        $response.c8yResponse.error | Should -Match "validationError"
        $response.message | Should -Match "Following mandatory fields should be included"
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
