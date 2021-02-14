. $PSScriptRoot/imports.ps1

Describe -Name "Invoke-ClientRequest" {

    It "gets a list of applications (defaults to GET method)" {
        $Response = Invoke-ClientRequest -Uri "/application/applications"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "should accept query parameters" {
        Invoke-ClientRequest -Uri "/alarm/alarms" -QueryParameters @{
                pageSize = "1";
            } `
            -Whatif `
            -InformationVariable requestInfo
        $LASTEXITCODE | Should -Be 0
        $requestInfo | Should -Not -BeNullOrEmpty
        ($requestInfo | Out-String) | Should -BeLike "*/alarm/alarms?pageSize=1*"
    }

    It "should return the raw json text when using -Raw" {
        $Response = Invoke-ClientRequest -Uri "/inventory/managedObjects" -QueryParameters @{
            pageSize = "2";
        } `
        -Raw
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Result = $Response | ConvertFrom-Json
        $Result.statistics | Should -Not -BeNullOrEmpty
        $Result.next | Should -Not -BeNullOrEmpty
        $Result.self | Should -Not -BeNullOrEmpty
    }

    It "should return the array of managed objects and not the raw response when not using -Raw" {
        $Response = Invoke-ClientRequest -Uri "/inventory/managedObjects" -QueryParameters @{
            pageSize = "2";
        }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Results = $Response | ConvertFrom-Json
        $Results | Should -HaveCount 2
    }

    It "should accept query parameters and return the raw response" {
        $options = @{
            Uri = "/alarm/alarms"
            QueryParameters = @{
                pageSize = "1";
            }
            WhatIf = $true
            InformationVariable = "requestInfo"
        }
        Invoke-ClientRequest @options
        $LASTEXITCODE | Should -Be 0
        $requestInfo | Should -Not -BeNullOrEmpty
        ($requestInfo | Out-String) | Should -BeLike "*/alarm/alarms?pageSize=1*"
    }

    It "post an inventory managed object from a string" {
        $Response = Invoke-ClientRequest -Uri "/inventory/managedObjects" -Method "post" -Data "name=test"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $obj = $Response | ConvertFrom-Json
        $obj.name | Should -BeExactly "test"

        if ($obj.id) {
            Remove-ManagedObject -Id $obj.id
        }
    }

    It "post an inventory managed object from a string with pretty print" {
        $Response = Invoke-ClientRequest `
            -Uri "/inventory/managedObjects" `
            -Method "post" `
            -Data "name=test" `
            -Pretty

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        ($Response -join "`n") | Should -BeLikeExactly '*"name": "test"*' -Because "Pretty print should have a space after the ':'"

        $obj = $Response | ConvertFrom-Json
        if ($obj.id) {
            Remove-ManagedObject -Id $obj.id
        }
    }

    It "Uploads a file to the inventory api" {
        $Text = "äüöp01!"
        $TestFile = New-TestFile -InputObject $Text
        $Response = Invoke-ClientRequest -Uri "inventory/binaries" -Method "post" -InFile $TestFile
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $obj = $Response | ConvertFrom-Json
        $obj.name | Should -BeExactly (Get-Item $TestFile).Name

        # Download file
        $BinaryContents = Get-Binary -Id $obj.id
        $BinaryContents | Should -BeExactly $Text

        # Cleanup
        Remove-Item $TestFile

        if ($obj.id) {
            Remove-Binary -Id $obj.id
        }
    }

    It "post an inventory managed object using uncompressed json text" {
        $jsonText = @"
{
    "name": "manual_object_001",
    "c8y_CustomObject": {
        "prop1": true
    }
}
"@
        $Response = Invoke-ClientRequest `
            -Uri "/inventory/managedObjects" `
            -Method "post" `
            -Data $jsonText

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $obj = $Response | ConvertFrom-Json
        $obj.name | Should -BeExactly "manual_object_001"

        if ($obj.id) {
            Remove-ManagedObject -Id $obj.id
        }
    }

    It "post an inventory managed object using hashtable" {
        $Response = Invoke-ClientRequest `
            -Uri "/inventory/managedObjects" `
            -Method "post" `
            -Data @{
                name = "manual_object_002"
                c8y_CustomObject = @{
                    prop1 = $false
                }
            }

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $obj = $Response | ConvertFrom-Json
        $obj.name | Should -BeExactly "manual_object_002"

        if ($obj.id) {
            Remove-ManagedObject -Id $obj.id
        }
    }

    It "Send request with custom headers" {
        Invoke-ClientRequest `
            -Uri "/inventory/managedObjects" `
            -Method "post" `
            -Headers @{
                MyHeader = "SomeValue"
                2 = 1
            } `
            -Data @{
                name = "manual_object_002"
                c8y_CustomObject = @{
                    prop1 = $false
                }
            } `
            -WhatIf `
            -InformationVariable requestInfo

        $LASTEXITCODE | Should -Be 0
        $requestInfo | Should -Not -BeNullOrEmpty

        ($requestInfo | Out-String) | Should -BeLike "*MyHeader: SomeValue*"
        ($requestInfo | Out-String) | Should -BeLike "*2: 1*"
    }

    It "Sends a request without a body" {
        Invoke-ClientRequest `
            -Uri "/inventory/managedObjects" `
            -Method "post" `
            -WhatIf `
            -InformationVariable requestInfo

        $LASTEXITCODE | Should -Be 0
        $requestInfo | Should -Not -BeNullOrEmpty
        ($requestInfo | Out-String) -match "Body:\s+\(empty\)" | Should -HaveCount 1
    }

    It "Sends a request with an empty hashtable" {
        Invoke-ClientRequest `
            -Uri "/inventory/managedObjects" `
            -Method "post" `
            -Data @{} `
            -WhatIf `
            -InformationVariable requestInfo

        $LASTEXITCODE | Should -Be 0
        $requestInfo | Should -Not -BeNullOrEmpty
        $requestInfo | Out-String | Should -match "(?ms)Body:\s*(\{.*\})"
    }

    It "Sends a request using templates" {
        $template = New-TemporaryFile
        @"
{
    c8y_CustomFragment: {
        test: true
    }
}
"@ |Out-File $template
        $options = @{
            Uri = "/inventory/managedObjects"
            Method = "post"
            Template = $template
        }
        $Response = Invoke-ClientRequest @options

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Result = $Response | ConvertFrom-Json
        $Result.c8y_CustomFragment.test | Should -BeExactly $true
        $Result.id | Remove-ManagedObject
    }

    It "Sends a request using deep nested object" {
        $data = @{
            root = @{
                level1 = @{
                    level2 = @{
                        level3 = @{
                            level4 = @{
                                level5 = @{
                                    value = 1
                                }
                            }
                        }
                    }
                }
            }
        }
        $Response = Invoke-ClientRequest `
            -Uri "/inventory/managedObjects" `
            -Method "post" `
            -Data $data

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Result = $Response | ConvertFrom-Json
        $Result.root.level1.level2.level3.level4.level5.value | Should -BeExactly 1
        $Result.id | Remove-ManagedObject
    }

    It "saves request information to the InformationVariable (hiding verbose messages)" {
        $response = Invoke-ClientRequest `
            -Uri "/inventory/managedObjects" `
            -Method "get" `
            -InformationVariable responseInfo

        $responseInfo | Should -Not -BeNullOrEmpty
        $responseInfo | Should -HaveCount 1
        $responseInfo.MessageData.request | Should -Not -BeNullOrEmpty
        $responseInfo.MessageData.requestHeader | Should -Not -BeNullOrEmpty
        $responseInfo.MessageData.responseHeader | Should -Not -BeNullOrEmpty
        $responseInfo.MessageData.responseTime | Should -Match "^\d+ms$"
        $responseInfo.MessageData.statusCode | Should -Match "^\d+$"
        $responseInfo.MessageData.responseLength | Should -Not -BeNullOrEmpty
    }
}
