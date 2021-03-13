. $PSScriptRoot/imports.ps1

Describe -Name "Inherit-Parameters" {
    BeforeAll {
        $env:C8Y_DISABLE_INHERITANCE = $null
    }

    BeforeEach {
        $ids = New-Object System.Collections.ArrayList
    }

    It "Force and WhatIf parameters are automatically inherited to module cmdlets" {
        Function Test-MyCustomFunction {
            [cmdletbinding()]
            Param(
                [switch] $Force
            )
            PSc8y\New-ManagedObject -Name "myname"
        }

        $mo = Test-MyCustomFunction -Force -WhatIf:$false
        $null = $ids.Add($mo.id)
        $LASTEXITCODE | Should -Be 0
        $mo | Should -Not -BeNullOrEmpty
        $mo.id | Should -Match "^\d+$"

        $mo = Test-MyCustomFunction -Force -WhatIf
        $mo | Should -BeNullOrEmpty
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
        $mo | Should -BeNullOrEmpty
        $LASTEXITCODE | Should -Be 0
    }

    It "WhatIf inheritance can be disabled via an environment variable allow user to control it themselves" {
        Function Test-MyCustomFunction {
            [cmdletbinding()]
            Param(
                [switch] $Force
            )
            PSc8y\New-ManagedObject -Name "myname" -Force:$Force
        }

        $env:C8Y_DISABLE_INHERITANCE = $true
        $WhatIfPreference = $null
        $mo = Test-MyCustomFunction -Force -WhatIf
        $ids.Add($mo.id)
        $mo | Should -Not -BeNullOrEmpty -Because "Function does not pass on WhatIf parameter"
        $LASTEXITCODE | Should -Be 0

        $WhatIfPreference = $true
        $mo = Test-MyCustomFunction -Force
        $ids.Add($mo.id)
        $mo | Should -Not -BeNullOrEmpty
        $LASTEXITCODE | Should -Be 0
    }

    AfterEach {
        $ids | Remove-ManagedObject
    }

    AfterAll {
        $env:C8Y_DISABLE_INHERITANCE = $null
    }
}
