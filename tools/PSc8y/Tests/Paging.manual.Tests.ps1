. $PSScriptRoot/imports.ps1

Describe -Name "Get-Pagination" {
    BeforeAll {
        $Device = New-TestDevice
        
        $Alarms = @(1..20) | ForEach-Object {
            PSc8y\New-Alarm `
                -Device $Device.id `
                -Type "c8y_TestAlarm$_" `
                -Time "-0s" `
                -Text "Test alarm $_" `
                -Severity MAJOR
        }

    }

    BeforeEach {
        $cliOutputFile = New-TemporaryFile
    }

    It "Get all of the alarms using IncludeAll" {
        $env:C8Y_SETTINGS_INCLUDEALL_PAGESIZE = "10"

        $Response = PSc8y\Get-AlarmCollection `
            -Device $Device.id `
            -IncludeAll `
            -Verbose 4> $cliOutputFile

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $VerboseOutput = Get-Content $cliOutputFile

        ($VerboseOutput -match "settings.includeAll.pageSize") | Should -BeLike "*settings.includeAll.pageSize: 10"

        # 2 because the first result does not have the "fetching next page"
        ($VerboseOutput -match "Fetching next page").Count | Should -BeExactly 2
    }

    AfterEach {
        if ($cliOutputFile -and (Test-Path $cliOutputFile)) {
            Remove-Item $cliOutputFile -Force
        }
    }

    AfterAll {
        PSc8y\Remove-ManagedObject -Id $Device.id

    }
}

