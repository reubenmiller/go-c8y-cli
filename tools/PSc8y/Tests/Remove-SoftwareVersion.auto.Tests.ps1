. $PSScriptRoot/imports.ps1

Describe -Name "Remove-SoftwareVersion" {
    BeforeEach {

    }

    It -Skip "Uninstall a software package version" {
        $Response = PSc8y\Remove-SoftwareVersion -Device $mo.id -Software go-c8y-cli -Version 1.0.0
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

