. $PSScriptRoot/imports.ps1

InModuleScope PSc8y {
    Describe -Name "New-MeasurementCollectionBatch" {
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

        It "TODO: Creates a collection of measurements using fixed data" {
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
                InputFile = $inputFile
                Workers = 1
                Data = $rawjson
                Template = "{time:time.now}"
                InformationVariable = "request"
            }
            # using WhatIf
            $options.WhatIf = $true
            $Response = New-MeasurementCollectionBatch @options
            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty
            $request -match "What If: Sending" | Should -HaveCount 2
            $request -match "Sending \[POST\] request to \[.*/measurement/measurements\]" | Should -HaveCount 2

            # send request
            $options.WhatIf = $false
            $Response = New-MeasurementCollectionBatch @options
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
                InputFile = $inputFile
                Workers = 1
                Template = $TemplateFile
                InformationVariable = "request"
            }

            $options.WhatIf = $true
            $Response = New-MeasurementCollectionBatch @options
            
            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty
            $request -match "What If: Sending" | Should -HaveCount 2
            $request -match "Sending \[POST\] request to \[.*/measurement/measurements\]" | Should -HaveCount 2

            # send request
            $options.WhatIf = $false
            $Response = New-MeasurementCollectionBatch @options
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
}