. $PSScriptRoot/imports.ps1

Describe -Name "Get-ManagedObjectCount" {
    BeforeEach {

    }

    It "Get count of managed objects" {
        $Response = PSc8y\Get-ManagedObjectCount
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

