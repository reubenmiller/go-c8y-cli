﻿. $PSScriptRoot/imports.ps1

Describe -Name "Invoke-ClientRequest" {

    It "gets a list of applications (defaults to GET method)" {
        $Response = Invoke-ClientRequest -Uri "/application/applications"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "should accept query parameters" {
        $output = Invoke-ClientRequest -Uri "/alarm/alarms" -QueryParameters @{
                pageSize = "1";
            } `
            -Dry -DryFormat json
        $LASTEXITCODE | Should -Be 0

        $requests = $output | ConvertFrom-Json
        $requests | Should -HaveCount 1
        $requests[0] | Should -MatchObject @{method = "GET"; pathEncoded = "/alarm/alarms?pageSize=1"} -Property method, pathEncoded
    }

    It "should return powershell objects by default" {
        $Response = Invoke-ClientRequest -Uri "/inventory/managedObjects" -QueryParameters @{
            pageSize = "2";
        }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response.managedObjects | Should -Not -BeNullOrEmpty
        $Response.statistics | Should -Not -BeNullOrEmpty
        $Response.next | Should -Not -BeNullOrEmpty
        $Response.self | Should -Not -BeNullOrEmpty
    }

    It "should return the raw json text when using" {
        $Response = Invoke-ClientRequest `
            -Uri "/inventory/managedObjects" -QueryParameters @{ pageSize = "2" } `
            -AsPSObject:$false
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
        $Response | Should -HaveCount 2
    }

    It "should accept query parameters and support Dry" {
        $options = @{
            Uri = "/alarm/alarms"
            QueryParameters = @{
                pageSize = "1";
            }
            Dry = $true
            DryFormat = "json"
        }
        $output = Invoke-ClientRequest @options
        $LASTEXITCODE | Should -Be 0

        $requests = $output | ConvertFrom-Json
        $requests | Should -HaveCount 1
        $requests[0] | Should -MatchObject @{method = "GET"; pathEncoded = "/alarm/alarms?pageSize=1"} -Property method, pathEncoded
    }

    It "return the raw response when a non-json accept header is used" {
        $testMeasurement = New-TestDevice | New-Measurement -Template "test.measurement.jsonnet"
        $dateFrom = c8y template execute --template "_.Now('-6h')"
        $Response = Invoke-ClientRequest -Uri "/measurement/measurements" -QueryParameters @{dateFrom = $dateFrom} -Method "get" -Accept "text/csv"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        Remove-Device $testMeasurement.source.id

        $Response | Should -HaveType string
    }

    It "post an inventory managed object from a string" {
        $Response = Invoke-ClientRequest -Uri "/inventory/managedObjects" -Method "post" -Data "name=test"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $obj = $Response
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
            -AsPSObject:$false `
            -Compact:$false -NoColor

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

        $obj = $Response
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

        $Response.name | Should -BeExactly "manual_object_001"

        if ($Response.id) {
            Remove-ManagedObject -Id $Response.id
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

        $Response.name | Should -BeExactly "manual_object_002"

        if ($Response.id) {
            Remove-ManagedObject -Id $Response.id
        }
    }

    It "Send request with custom headers" {
        $output = Invoke-ClientRequest `
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
            -DryFormat "json" `
            -Dry

        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty

        $request = $output | ConvertFrom-Json
        $request | Should -HaveCount 1
        $request.headers.MyHeader | Should -BeExactly "SomeValue"
        $request.headers.2 | Should -BeExactly "1"
    }

    It "Sends a request without a body" {
        $output = Invoke-ClientRequest `
            -Uri "/inventory/managedObjects" `
            -Method "post" `
            -Dry

        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
        ($output | Out-String) -match "Body:\s+\(empty\)" | Should -HaveCount 1
    }

    It "Sends a request with an empty hashtable" {
        $output = Invoke-ClientRequest `
            -Uri "/inventory/managedObjects" `
            -Method "post" `
            -Data @{} `
            -DryFormat "json" `
            -Dry

        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
        $request = $output | ConvertFrom-Json
        $request | Should -HaveCount 1
        ($request.body | ConvertTo-Json) | Should -BeExactly "{}"
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
        $Response.c8y_CustomFragment.test | Should -BeExactly $true
        $Response.id | Remove-ManagedObject
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
        $Response.root.level1.level2.level3.level4.level5.value | Should -BeExactly 1
        $Response.id | Remove-ManagedObject
    }

    It "uses existing environment variables for native powershell requests" {
        Function Invoke-MyRequest {
            [cmdletbinding()]
            Param(
                [string] $Path,

                [string] $Method = "GET",

                [object] $Body,

                [object] $Accept = "application/json",

                [object] $ContentType = "application/json",

                [hashtable] $Headers,

                # Additional options that will be passed to Invoke-RestMethod
                [hashtable] $AdditionalOptions
            )
            $options = @{
                Uri = "$env:C8Y_HOST".TrimEnd("/") + "/" + "$Path".TrimStart("/")
                Method = $Method
                ContentType = $ContentType
                Headers = @{ Authorization = $env:C8Y_HEADER_AUTHORIZATION }
            }
            if ($Accept) {
                $options.Headers.Accept = $Accept
            }
            if ($Headers) {
                $options.Headers += $Headers
            }
            if ($Body) {
                if ($Body -is [hashtable]) {
                    $options.Body = ConvertTo-Json $Body -Compress
                } else {
                    $options.Body = $Body
                }
            }
            if ($AdditionalOptions) {
                $options += $AdditionalOptions
            }

            Invoke-RestMethod @options
        }

        # Force loading session info as environment variables
        c8y sessions set --session $env:C8Y_SESSION --shell auto | Out-String | Invoke-Expression

        $output = Invoke-MyRequest -Path "user/currentUser"
        $output | Should -Not -BeNullOrEmpty

        $output = Invoke-MyRequest "/inventory/managedObjects" -Method "POST" -Body @{
            name = "testme"
        }
        $output.name | Should -Be "testme"
        $output.id | Remove-ManagedObject
    }
}
