. $PSScriptRoot/imports.ps1

Describe -Name "New-Software" {
    BeforeEach {
        $type = New-RandomString -Prefix "customType_"

    }

    It "Create a managed object" {
        $Response = PSc8y\New-ManagedObject -Name "python3-requests" -Description "python requests library" -Data @{$type=@{}}
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-ManagedObjectCollection -FragmentType $type | Remove-ManagedObject

    }
}

