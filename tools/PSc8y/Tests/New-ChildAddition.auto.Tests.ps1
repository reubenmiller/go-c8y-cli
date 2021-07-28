. $PSScriptRoot/imports.ps1

Describe -Name "New-ChildAddition" {
    BeforeEach {
        $software = PSc8y\New-ManagedObject -Name softwarePackage1

    }

    It "Create a child addition and link it to an existing managed object" {
        $Response = PSc8y\New-ChildAddition -Id $software.id -Data "custom.value=test" -Global
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $software.id

    }
}

