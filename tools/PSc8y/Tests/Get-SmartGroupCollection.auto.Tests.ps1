. $PSScriptRoot/imports.ps1

Describe -Name "Get-SmartGroupCollection" {
    BeforeEach {
        $SmartGroup1 = New-TestSmartGroup
        $Name = $SmartGroup1.name

    }

    It "Get a list of smart groups" {
        $Response = PSc8y\Get-SmartGroupCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a list of smart groups with the names starting with 'myText'" {
        $Response = PSc8y\Get-SmartGroupCollection -Name "$Name*"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $SmartGroup1.id

    }
}

