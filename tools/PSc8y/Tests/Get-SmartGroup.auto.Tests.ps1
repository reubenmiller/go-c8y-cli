. $PSScriptRoot/imports.ps1

Describe -Name "Get-SmartGroup" {
    BeforeEach {
        $smartgroup = PSc8y\New-TestSmartGroup

    }

    It "Get smart group by id" {
        $Response = PSc8y\Get-SmartGroup -Id $smartgroup.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get smart group by name" {
        $Response = PSc8y\Get-SmartGroup -Id $smartgroup.name
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $smartgroup.id

    }
}

