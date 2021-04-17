. $PSScriptRoot/imports.ps1

Describe -Name "Update-SmartGroup" {
    BeforeEach {
        $smartgroup = PSc8y\New-TestSmartGroup

    }

    It "Update smart group by id" {
        $Response = PSc8y\Update-SmartGroup -Id $smartgroup.id -NewName "MyNewName"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update smart group by name" {
        $Response = PSc8y\Update-SmartGroup -Id $smartgroup.name -NewName "MyNewName"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update smart group custom properties" {
        $Response = PSc8y\Update-SmartGroup -Id $smartgroup.name -Data @{ "myValue" = @{ value1 = $true } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $smartgroup.id

    }
}

