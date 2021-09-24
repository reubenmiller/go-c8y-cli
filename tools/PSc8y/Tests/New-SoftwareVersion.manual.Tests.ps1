. $PSScriptRoot/imports.ps1

Describe -Name "New-SoftwareVersion" {
    BeforeEach {
        $name = New-RandomString -Prefix "customPackage_"
        $software = New-Software -Name $name -Description "Example package"

    }

    It "Create a new version to an existing software package" {
        $Response = PSc8y\New-SoftwareVersion -Software $name -Version "1.2.3" -Url "https://test.com/blobs/mypackage.deb"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-Software -Id $software.id

    }
}

