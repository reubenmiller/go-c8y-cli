. $PSScriptRoot/imports.ps1

InModuleScope PSc8y {
    Describe -Skip -Tag "Deprecated" -Name "Add-DeviceCollectionToGroupBatch" {
        BeforeEach {
            $ids = New-Object System.Collections.ArrayList
            $inputFile = New-TemporaryFile

            $group = New-TestDeviceGroup
            $devices = @(1..5) | ForEach-Object {
                New-TestDevice
            }
            $null = $ids.Add($group.id)
            $null = $ids.AddRange($devices.id)

            # Save ids to file
            $devices.id | Out-File $inputFile
        }

        It "Adds a managed objects to a group via a file containing managed objects ids" {
            $options = @{
                Group = $group
                InputFile = $inputFile
                Delay = 1000
                Workers = 5
                InformationVariable = "Request"
                ErrorVariable = "ErrorMessages"
            }

            # WhatIf
            $options.WhatIf = $true
            $Response = Add-DeviceCollectionToGroupBatch @options
            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty
            $Request -match "What If: Sending \[POST\] request to \[.*/inventory/managedObjects/\d+/childAssets\]" | Should -HaveCount 5
            $children = Get-ChildAssetCollection -Group $group.id -PageSize 10
            $children | Should -BeNullOrEmpty

            # Real request
            $options.WhatIf = $false
            $Response = Add-DeviceCollectionToGroupBatch @options
            $LASTEXITCODE | Should -Be 0
            $Response | Should -Not -BeNullOrEmpty
            $Response | Should -HaveCount 5
            $children = Get-ChildAssetCollection -Group $group.id -PageSize 10
            $children | Should -HaveCount 5
            ($children.id | Sort-Object) | Should -Be -ExpectedValue ($devices.id | Sort-Object)

        }

        AfterEach {
            $ids | Remove-ManagedObject
            if (Test-Path $inputFile) {
                Remove-Item $inputFile
            }
        }
    }
}