. $PSScriptRoot/imports.ps1

Describe -Name "Remove-Software" {
    BeforeEach {
        $mo = PSc8y\New-Software -Name "python3-requests"
        $mo = PSc8y\New-ManagedObject -Name "testMO"

    }

    It "Delete a software package and all related versions" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Remove-Software -ForceCascade:$false
        $LASTEXITCODE | Should -Be 0
    }

    It "Delete a software package (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $mo.id | Remove-Software
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-ManagedObject -Id $mo.id -ErrorAction SilentlyContinue

    }
}

