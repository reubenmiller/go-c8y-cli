. $PSScriptRoot/imports.ps1

Describe -Name "Get-BulkOperationCollection" {
    BeforeEach {

    }

    It "Get a list of bulk operations" {
        $Response = PSc8y\Get-BulkOperationCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

