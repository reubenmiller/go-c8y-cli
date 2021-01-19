. $PSScriptRoot/imports.ps1

Describe -Name "Get-ChildAdditionCollection" {
    BeforeEach {
        $software = PSc8y\New-ManagedObject -Name softwarePackage1
        $version = PSc8y\New-ManagedObject -Name softwareVersion1
        PSc8y\Add-ChildAddition -Id $software.id -NewChild $version.id

    }

    It "Get a list of the child additions of an existing managed object" {
        $Response = PSc8y\Get-ChildAdditionCollection -Id $software.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a list of the child additions of an existing managed object (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $software.id | Get-ChildAdditionCollection
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-ManagedObject -Id $software.id
        PSc8y\Remove-ManagedObject -Id $version.id

    }
}

