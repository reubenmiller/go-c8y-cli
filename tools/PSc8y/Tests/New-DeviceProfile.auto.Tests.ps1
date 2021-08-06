. $PSScriptRoot/imports.ps1

Describe -Name "New-DeviceProfile" {
    BeforeEach {
        $type = New-RandomString -Prefix "customType_"

    }

    It "Create a managed object" {
        $Response = PSc8y\New-ManagedObject -Name "python3-requests" -Data @{$type=@{}}
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-ManagedObjectCollection -FragmentType $type | Remove-ManagedObject

    }
}

