. $PSScriptRoot/imports.ps1

Describe -Name "Get-SoftwareVersion" {
    BeforeEach {
        $software = PSc8y\New-Software -Name "testMO"
        $softwareVersion = PSc8y\New-SoftwareVersion -Software $software.id -Version "1.0.0" -Url "https://test.com/file.mender"

    }

    It "Get a software package version using name" {
        $Response = PSc8y\Get-SoftwareVersion -Id $softwareVersion.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Get a software package version (using pipeline)" {
        $Response = PSc8y\Get-ManagedObject -Id $softwareVersion.id | Get-SoftwareVersion
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Software -Id $software.id -ErrorAction SilentlyContinue

    }
}

