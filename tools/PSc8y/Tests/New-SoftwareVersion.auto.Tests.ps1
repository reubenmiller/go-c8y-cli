. $PSScriptRoot/imports.ps1

Describe -Name "New-SoftwareVersion" {
    BeforeEach {
        $type = New-RandomString -Prefix "customType_"

    }

    It "Create a new version to an existing software package" {
        $Response = PSc8y\New-ManagedObject -Name "python3-requests" -Description "python requests library" -Data @{$type=@{}}
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-ManagedObjectCollection -FragmentType $type | Remove-ManagedObject

    }
}

