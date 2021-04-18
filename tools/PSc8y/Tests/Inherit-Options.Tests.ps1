. $PSScriptRoot/imports.ps1

Describe -Name "Inherit-Parameters" {
    BeforeAll {
        $env:C8Y_DISABLE_INHERITANCE = $null
    }

    BeforeEach {
        $ids = New-Object System.Collections.ArrayList
    }

    It "Force and Dry parameters are automatically inherited to module cmdlets" {
        Function Test-MyCustomFunction {
            [cmdletbinding()]
            Param()
            DynamicParam {
                Get-ClientCommonParameters -Type "Create", "Template"
            }
            Process {
                $options = @{ name = "myname" } + $PSBoundParameters
                PSc8y\New-ManagedObject @options
            }
        }

        $mo = Test-MyCustomFunction -Force -Dry:$false
        $null = $ids.Add($mo.id)
        $LASTEXITCODE | Should -Be 0
        $mo | Should -Not -BeNullOrEmpty
        $mo.id | Should -Match "^\d+$"

        $mo = Test-MyCustomFunction -Force -Dry
        $mo | Should -Not -BeNullOrEmpty
        $requests = $mo | ConvertFrom-Json
        $requests.path | Should -BeExactly "/inventory/managedObjects"
        $LASTEXITCODE | Should -Be 0

        $VerboseMessage = $($mo = Test-MyCustomFunction -Force -Verbose) 2>&1
        $mo.id | Should -Match "^\d+$"
        $null = $ids.Add($mo.id)
        $mo | Should -Not -BeNullOrEmpty
        $LASTEXITCODE | Should -Be 0
        $VerboseMessage | Should -Not -BeNullOrEmpty

        $VerboseMessage = $($mo = Test-MyCustomFunction -Force -Verbose:$false) 2>&1
        $mo.id | Should -Match "^\d+$"
        $null = $ids.Add($mo.id)
        $mo | Should -Not -BeNullOrEmpty
        $LASTEXITCODE | Should -Be 0
        $VerboseMessage | Should -BeNullOrEmpty

        $VerboseMessage = $($mo = Test-MyCustomFunction -Force) 2>&1
        $mo.id | Should -Match "^\d+$"
        $null = $ids.Add($mo.id)
        $mo | Should -Not -BeNullOrEmpty
        $LASTEXITCODE | Should -Be 0
        $VerboseMessage | Should -BeNullOrEmpty

        # Set via preference
        $VerbosePreference = "Continue"
        $VerboseMessage = $($mo = Test-MyCustomFunction -Force) 2>&1
        $mo.id | Should -Match "^\d+$"
        $null = $ids.Add($mo.id)
        $mo | Should -Not -BeNullOrEmpty
        $LASTEXITCODE | Should -Be 0
        $VerboseMessage | Should -Not -BeNullOrEmpty

        # Reset verbose preference
        $VerbosePreference = ""

        $VerboseMessage = $($mo = Test-MyCustomFunction -Force) 2>&1
        $mo.id | Should -Match "^\d+$"
        $null = $ids.Add($mo.id)
        $mo | Should -Not -BeNullOrEmpty
        $LASTEXITCODE | Should -Be 0
        $VerboseMessage | Should -BeNullOrEmpty

        $WhatIfPreference = $true
        $mo = Test-MyCustomFunction -Force
        $mo | Should -Not -BeNullOrEmpty
        $LASTEXITCODE | Should -Be 0
        $requests = $mo | ConvertFrom-Json
        $requests.path | Should -BeExactly "/inventory/managedObjects"
    }

    AfterEach {
        $ids | Remove-ManagedObject
    }

    AfterAll {
        $env:C8Y_DISABLE_INHERITANCE = $null
    }
}
