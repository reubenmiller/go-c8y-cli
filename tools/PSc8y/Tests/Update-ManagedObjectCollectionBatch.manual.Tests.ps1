. $PSScriptRoot/imports.ps1

InModuleScope PSc8y {
    Describe -Skip -Name "Update-ManagedObject" {
        BeforeEach {
            $ids = New-Object System.Collections.ArrayList
            $type = New-RandomString -Prefix "customType_"
            $inputFile = New-TemporaryFile

            $Device1 = New-TestDevice
            $Device2 = New-TestDevice
            $null = $ids.AddRange(@($Device1.id, $Device2.id))

            # Save ids to file
            $ids | Out-File $inputFile
            
        }

        It "Updates a collection of managed objects using fixed data" {
            $rawjson = @"
{
    "type": "$type",
    "c8y_Kpi": {
        "max": 19.1010101E19,
        "description": ""
    }
}
"@
            # base options
            $options = @{
                Workers = 1
                Count = 2
                Data = $rawjson
                InformationVariable = "request"
            }

            # using WhatIf
            $options.WhatIf = $true
            $Response = Update-ManagedObjectCollectionBatch @options
            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty
            $request -match "What If: Sending" | Should -HaveCount 2
            $request -match "Sending \[POST\] request to \[.*/inventory/managedObjects\]" | Should -HaveCount 2

            # send request
            $options.WhatIf = $false
            $Response = Update-ManagedObjectCollectionBatch @options
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response | Should -HaveCount 2
        }

        It "Updates a managed object from a json file" {
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

            $Response = Update-ManagedObjectCollectionBatch -Workers 1 -Count 2 -Data $jsonfile -WhatIf -InformationVariable request
            Remove-Item $jsonfile -Force
            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty
            $request -match "What If: Sending" | Should -HaveCount 2
            $request -match "Sending \[POST\] request to \[.*/inventory/managedObjects\]" | Should -HaveCount 2
        }

        It "Throws an error if the json file contains invalid json" {
            $jsonfile = New-TemporaryFile
            '{"name": ' | Out-File $jsonfile

            $Response = Update-ManagedObjectCollectionBatch -Workers 1 -AbortOnErrors 1 -Count 2 -Data $jsonfile -WhatIf -InformationVariable request
            Remove-Item $jsonfile -Force

            $LASTEXITCODE | Should -Not -Be 0
            $Response | Should -BeNullOrEmpty
            $request -match "What If: Sending" | Should -HaveCount 0
            $request -match "Sending \[POST\] request to \[.*/inventory/managedObjects\]" | Should -HaveCount 0
        }

        It "Updates a collection of managed objects using a jsonnet template" {
            $template = New-TemporaryFile
            @"
{
    name: "object" + rand.index,
    type: "$type",
    my_Complex: {
        values: [1,2,3,4]
    }
}
"@ | Out-File $template

            # using WhatIf
            $Response = Update-ManagedObjectCollectionBatch -Workers 1 -Count 2 -Template $template -WhatIf -InformationVariable request
            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty
            $request -match "What If: Sending" | Should -HaveCount 2
            $request -match "Sending \[POST\] request to \[.*/inventory/managedObjects\]" | Should -HaveCount 2

            # send request
            $Response = Update-ManagedObjectCollectionBatch -Workers 5 -Count 2 -Template $template -InformationVariable request
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response | Should -HaveCount 2
            # The order of response is only known if Workers is set to 1!
            $Response.name -eq "object1" | Should -HaveCount 1
            $Response.name -eq "object2" | Should -HaveCount 1
        }

        It "Aborts if the jsonet template is invalid" {
            $template = New-TemporaryFile
            @"
local +123+123
{
    name: "object" + rand.index,
    type: "$type",
    my_Complex: {
        values: [1,2,3,4]
    }
}
"@ | Out-File $template

            # using WhatIf
            $Response = Update-ManagedObjectCollectionBatch -Workers 1 -AbortOnErrors 1 -Count 2 -Template $template -WhatIf -InformationVariable request
            $LASTEXITCODE | Should -Be 103
            $Response | Should -BeNullOrEmpty

            # send request
            $Response = Update-ManagedObjectCollectionBatch -Workers 5 -AbortOnErrors 1 -Count 2 -Template $template -InformationVariable request
            $LASTEXITCODE | Should -Be 103
            $Response | Should -BeNullOrEmpty

            # send request (allow to complete, but with errors)
            $Response = Update-ManagedObjectCollectionBatch -Workers 5 -AbortOnErrors 3 -Count 2 -Template $template -InformationVariable request
            $LASTEXITCODE | Should -Be 104
            $Response | Should -BeNullOrEmpty
        }

        It "Managed object allow setting the processing mode" {
            foreach ($mode in @("PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP")) {
                $options = @{
                    Data = @{}
                    ProcessingMode = $mode
                    Workers = 5
                    Count = 1
                    WhatIf = $true
                    InformationVariable = "Request"
                }
                $Response = Update-ManagedObjectCollectionBatch @options
                $LASTEXITCODE | Should -Be 0
                $Response | Should -BeNullOrEmpty
                ($Request | Out-String) -match "X-Cumulocity-Processing-Mode:\s+$mode" | Should -HaveCount 1
            }
        }

        It "throws an error when then worker number exceeds the allowed value" {
            $options = @{
                Workers = 20
                Count = 2
                Data = @{name="test"}
                WhatIf = $true
                InformationVariable = "Request"
                ErrorVariable = "ErrorMessages"
            }
            $Response = Update-ManagedObjectCollectionBatch @options
            $LASTEXITCODE | Should -Be 100
            $Response | Should -BeNullOrEmpty
            $Request | Should -BeNullOrEmpty
            $ErrorMessages[-1] | Should -Match "number of workers exceeds"
        }

        AfterEach {
            Get-ManagedObjectCollection -Type $type -PageSize 100 | Select-Object | Remove-ManagedObject
            $ids | Remove-ManagedObject
        }
    }
}