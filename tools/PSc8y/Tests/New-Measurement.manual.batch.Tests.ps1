. $PSScriptRoot/imports.ps1

Describe -Name "New-Measurement" {
    BeforeEach {
        $ids = New-Object System.Collections.ArrayList
        $inputFile = New-TemporaryFile
        $TemplateFile = New-TemporaryFile

        $devices = @(1..2) | ForEach-Object {
            New-TestDevice
        }
        $null = $ids.AddRange($devices.id)

        # Save ids to file
        $devices.id | Out-File $inputFile
    }

    It "Creates a collection of measurements using fixed data" {
        # How to inject now into data
        $rawjson = @"
{
    "type": "myType",
    "c8y_Sensors": {
        "outsideTemperature": {
            "value": 30.1,
            "unit": "°C"
        }
    }
}
"@
        $options = @{
            Workers = 1
            Data = $rawjson
            Template = "{time:_.Now()}"
        }
        # using Dry
        $options.Dry = $true
        $options.DryFormat = "json"
        $output = $( $Response = Get-Content $inputFile | batch | New-Measurement @options ) 2>&1
        $LASTEXITCODE | Should -Be 0
        
        $requests = $output | ConvertFrom-Json
        $requests | Should -HaveCount 2
        $requests[0] | Should -MatchObject @{method = "POST"; pathEncoded = "/measurement/measurements"} -Property method, pathEncoded
        $requests[1] | Should -MatchObject @{method = "POST"; pathEncoded = "/measurement/measurements"} -Property method, pathEncoded

        # send request
        $options.Dry = $false
        $Response = Get-Content $inputFile | batch | New-Measurement @options
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response | Should -HaveCount 2
    }

    It "Creates a collection of measurements on devices using a template" {

        @"
{
"time": time.now,
"type": "myType",
"c8y_MyCustomFragment": {},
"c8y_Sensors": {
    "outsideTemperature": {
        "value": 30.1,
        "unit": "°C"
    }
}
}
"@ | Out-File $TemplateFile

        $options = @{
            Workers = 1
            Template = $TemplateFile
        }

        $options.Dry = $true
        $options.DryFormat = "json"
        $output = $( $Response = Get-Content $inputFile | batch | New-Measurement @options ) 2>&1
        
        $LASTEXITCODE | Should -Be 0
        $requests = $output | ConvertFrom-Json
        $requests | Should -HaveCount 2
        $requests[0] | Should -MatchObject @{method = "POST"; pathEncoded = "/measurement/measurements"} -Property method, pathEncoded
        $requests[1] | Should -MatchObject @{method = "POST"; pathEncoded = "/measurement/measurements"} -Property method, pathEncoded


        # send request
        $options.Dry = $false
        $Response = Get-Content $inputFile | batch | New-Measurement @options
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response | Should -HaveCount 2
    }

    AfterEach {
        $ids | Remove-ManagedObject
        if (Test-Path $inputFile) {
            Remove-Item $inputFile
        }
        Remove-Item $TemplateFile -Force
    }
}
