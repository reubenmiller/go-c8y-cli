. $PSScriptRoot/imports.ps1

Describe -Name "New-Configuration" {
    BeforeEach {
        $type = New-RandomString -Prefix "customType_"

    }

    It "Create a new configuration file" {
        $Response = PSc8y\New-Configuration -Name "agent config" -Description "Default agent configuration" -ConfigurationType "agentConfig" -Url "https://test.com/content/raw/app.json" -Data @{$type=@{}}
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-ManagedObjectCollection -FragmentType $type | Remove-ManagedObject

    }
}

