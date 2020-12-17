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

    It "Creates a managed object from a json file" {
        $jsonfile = New-TemporaryFile
        @"
{
    "name": "testMO",
    "type": "$type",
    "c8y_SoftwareList": [
        { "name": "app1", "version": "1.0.0", "url": "https://example.com/myfile1.deb"},
        { "name": "app2", "version": "9", "url": "https://example.com/myfile1.deb"},
        { "name": "app3 test", "version": "1.1.1", "url": "https://example.com/myfile1.deb"}
    ]
}
"@ | Out-File $jsonfile

        $Response = PSc8y\New-ManagedObject -Data $jsonfile.FullName
        Remove-Item $jsonfile -Force

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response.id | Should -MatchExactly "^\d+$"
        $Response.type | Should -BeExactly $type
        $Response.name | Should -BeExactly "testMO"

        $Response.c8y_SoftwareList[0].name | Should -BeExactly "app1"
        $Response.c8y_SoftwareList[1].name | Should -BeExactly "app2"
        $Response.c8y_SoftwareList[2].name | Should -BeExactly "app3 test"
    }

    It "Throws an error if the json file contains invalid json" {
        $jsonfile = New-TemporaryFile
        '{"name": ' | Out-File $jsonfile

        $Response = PSc8y\New-ManagedObject -Data $jsonfile.FullName
        Remove-Item $jsonfile -Force

        $LASTEXITCODE | Should -Not -BeExactly 0
        $Response | Should -BeNullOrEmpty
    }

    AfterEach {
        Get-ManagedObjectCollection -Type $type -PageSize 100 | Select-Object | Remove-ManagedObject

    }
}
