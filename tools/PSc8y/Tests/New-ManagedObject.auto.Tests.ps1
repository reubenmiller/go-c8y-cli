. $PSScriptRoot/imports.ps1

Describe -Name "New-ManagedObject" {
    BeforeEach {
        $type = New-RandomString -Prefix "customType_"

    }

    It "Create a managed object" {
        $Response = PSc8y\New-ManagedObject -Name "testMO" -Type $type -Data @{ custom_data = @{ value = 1 } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-ManagedObjectCollection -Type $type | Remove-ManagedObject

    }
}

