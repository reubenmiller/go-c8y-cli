. $PSScriptRoot/imports.ps1

Describe -Name "New-ManagedObject" {
    BeforeEach {
        $type = New-RandomString -Prefix "customType_"
    }

    It "Create a managed object with json in exponention notation" {
        $rawjson = @"
{
    "type": "",
    "c8y_Kpi": {
        "max": 19.1010101E19,
        "description": ""
    }
}
"@

        $Response = PSc8y\New-ManagedObject -Name "testMO" -Type $type -Data $rawjson
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create a managed object with escaped double quotes" {
        $rawjson = @"
{
    "type": "",
    "c8y_Kpi": {
        "description": "some \"value\" ok"
    }
}
"@

        $Response = PSc8y\New-ManagedObject -Name "testMO" -Type $type -Data $rawjson
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    AfterEach {
        Get-ManagedObjectCollection -Type $type | Select-Object | Remove-ManagedObject

    }
}
