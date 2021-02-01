. $PSScriptRoot/imports.ps1

InModuleScope PSc8y {
    Describe -Skip -Name "Remove-ManagedObjectCollectionBatch" {
        BeforeEach {
            $ids = New-Object System.Collections.ArrayList
            $inputFile = New-TemporaryFile
            $type = New-RandomString -Prefix "batch"

            $managedObjects = @(1..5) | ForEach-Object {
                New-ManagedObject -Type $type
            }
            $null = $ids.AddRange($managedObjects.id)

            # Save ids to file
            $managedObjects.id | Out-File $inputFile
        }

        It "Remove a list of managed objects via a file containing managed objects ids" {
            $options = @{
                InputFile = $inputFile
                Delay = 1000
                Workers = 5
                InformationVariable = "Request"
                ErrorVariable = "ErrorMessages"
            }

            # WhatIf
            $options.WhatIf = $true
            $Response = Remove-ManagedObjectCollectionBatch @options
            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty
            $Request -match "What If: Sending \[DELETE\] request to \[.*/inventory/managedObjects/\d+\]" | Should -HaveCount 5

            # Real request
            $options.WhatIf = $false
            $Response = Remove-ManagedObjectCollectionBatch @options
            $LASTEXITCODE | Should -Be 0
            $Response | Should -BeNullOrEmpty

            # Check if objects are really deleted
            $AfterDeletion = $managedObjects.id | Get-ManagedObject -ErrorAction SilentlyContinue
            $AfterDeletion | Should -BeNullOrEmpty
        }

        AfterEach {
            $ids | Remove-ManagedObject -ErrorAction SilentlyContinue
            if (Test-Path $inputFile) {
                Remove-Item $inputFile
            }
        }
    }
}