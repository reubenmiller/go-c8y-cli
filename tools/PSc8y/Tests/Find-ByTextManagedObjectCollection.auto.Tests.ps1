. $PSScriptRoot/imports.ps1

Describe -Name "Find-ByTextManagedObjectCollection" {
    BeforeEach {
        $Device1 = New-TestDevice

    }

    It "Find a list of managed objects by text" {
        $Response = PSc8y\Find-ByTextManagedObjectCollection -Text $Device1.name
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Find managed objects which contain the text 'myText' (using pipeline)" {
        $Response = PSc8y\Find-ByTextManagedObjectCollection -Text $Device1.name
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device1.id

    }
}

