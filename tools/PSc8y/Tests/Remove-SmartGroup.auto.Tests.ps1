. $PSScriptRoot/imports.ps1

Describe -Name "Remove-SmartGroup" {
    BeforeEach {
        $smartgroup = PSc8y\New-TestSmartGroup

    }

    It "Remove smart group by id" {
        $Response = PSc8y\Remove-SmartGroup -Id $smartgroup.id
        $LASTEXITCODE | Should -Be 0
    }

    It "Remove smart group by name" {
        $Response = PSc8y\Remove-SmartGroup -Id $smartgroup.name
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {

    }
}

