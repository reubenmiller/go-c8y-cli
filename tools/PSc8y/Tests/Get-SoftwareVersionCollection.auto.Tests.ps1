. $PSScriptRoot/imports.ps1

Describe -Name "Get-SoftwareVersionCollection" {
    BeforeEach {
        $software = PSc8y\New-Software -Name "testMO"
        $softwareVersion = PSc8y\New-SoftwareVersion -Software $software.id -Version "1.0.0" -Url "https://test.com/file.mender"

    }

    It "Get a list of software package versions" {
        $Response = PSc8y\Get-SoftwareVersionCollection -Software $software.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Software -Id $software.id -ErrorAction SilentlyContinue

    }
}

