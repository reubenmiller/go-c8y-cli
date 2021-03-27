. $PSScriptRoot/imports.ps1

Describe -Name "Get-ChildAssetCollection" {
    BeforeEach {

    }

    It -Skip "Get a list of the child assets of an existing device" {
        $Response = PSc8y\Get-ChildAssetCollection -Id 12345
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

