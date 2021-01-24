. $PSScriptRoot/imports.ps1

Describe -Name "Add-ChildAddition" {
    BeforeEach {
        $software = PSc8y\New-ManagedObject -Name softwarePackage1
        $version = PSc8y\New-ManagedObject -Name softwareVersion1

    }

    It "Add a related managed object as a child to an existing managed object" {
        $Response = PSc8y\Add-ChildAddition -Id $software.id -NewChild $version.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $software.id
        PSc8y\Remove-ManagedObject -Id $version.id

    }
}

