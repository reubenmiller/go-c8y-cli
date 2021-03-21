. $PSScriptRoot/imports.ps1

Describe -Name "Get-Pagination" {
    BeforeAll {
        $backupEnvSettings = @{
            C8Y_SETTINGS_INCLUDEALL_PAGESIZE = $env:C8Y_SETTINGS_INCLUDEALL_PAGESIZE
            C8Y_SETTINGS_DEFAULTS_PAGESIZE = $env:C8Y_SETTINGS_DEFAULTS_PAGESIZE
        }
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

    It "Get all of the alarms using IncludeAll and custom include all page size" {
        $env:C8Y_SETTINGS_INCLUDEALL_PAGESIZE = "10"

        $Response = PSc8y\Get-AlarmCollection `
            -Device $Device.id `
            -IncludeAll `
            -Debug 2> $cliOutputFile
        $C8Y_SETTINGS_INCLUDEALL_PAGESIZE = ""

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response | Should -HaveCount 20

        $VerboseOutput = Get-Content $cliOutputFile

        ($VerboseOutput -match "pageSize=10") | Should -Not -BeNullOrEmpty

        # 2 because the first result does not have the "fetching next page"
        ($VerboseOutput -match "Fetching next page").Count | Should -BeExactly 2
    }

    It "Get all of the alarms using IncludeAll and uneven custom include size" {
        $env:C8Y_SETTINGS_INCLUDEALL_PAGESIZE = "12"

        $Response = PSc8y\Get-AlarmCollection `
            -Device $Device.id `
            -IncludeAll `
            -Debug 2> $cliOutputFile

        $LASTEXITCODE | Should -Be 0
        $C8Y_SETTINGS_INCLUDEALL_PAGESIZE = ""
        $Response | Should -Not -BeNullOrEmpty
        $Response | Should -HaveCount 20

        $VerboseOutput = Get-Content $cliOutputFile

        ($VerboseOutput -match "pageSize=12") | Should -Not -BeNullOrEmpty

        # 1 because only one extra fetch should be required
        # as the first has 12 results, and the second result set has less than the requested
        # page size, so it should not try to fetch another page
        ($VerboseOutput -match "Fetching next page").Count | Should -BeExactly 1
    }

    It "Using include All with Dry" {
        $env:C8Y_SETTINGS_INCLUDEALL_PAGESIZE = ""

        $output = PSc8y\Get-DeviceCollection `
            -IncludeAll `
            -Dry `
            -DryFormat json 2>&1

        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
        $requests = $output | ConvertFrom-Json

        $requests[0].query | Should -Match "pageSize=2000"
    }

    It "Set default pagesize using environment setting" {
        $env:C8Y_SETTINGS_DEFAULTS_PAGESIZE = "10"

        $output = PSc8y\Get-AlarmCollection `
            -Device $Device.id `
            -Dry `
            -DryFormat json 2>&1

        $LASTEXITCODE | Should -Be 0
        $C8Y_SETTINGS_DEFAULTS_PAGESIZE = ""
        $output | Should -Not -BeNullOrEmpty
        $requests = $output | ConvertFrom-Json

        $requests | Should -HaveCount 1
        $requests[0].query | Should -Match "pageSize=10"
    }

    It "All collection commands support paging parameters" {
        $ExcludeCmdlets = @(
            "Get-SessionCollection",
            "Get-CurrentTenantApplicationCollection"
        )
        $cmdlets = Get-Command -Module PSc8y -Name "Get-*Collection*" |
            Where-Object {
                $ExcludeCmdlets -notcontains $_.Name
            }

        foreach ($icmdlet in $cmdlets) {
            $icmdlet | Should -HaveParameter "CurrentPage"
            $icmdlet | Should -HaveParameter "TotalPages"
            $icmdlet | Should -HaveParameter "IncludeAll"
            $icmdlet | Should -HaveParameter "PageSize"
            $icmdlet | Should -HaveParameter "WithTotalPages"
            $icmdlet | Should -HaveParameter "Raw"
        }
    }

    AfterEach {
        if ($cliOutputFile -and (Test-Path $cliOutputFile)) {
            Remove-Item $cliOutputFile -Force
        }
    }

    AfterAll {
        PSc8y\Remove-ManagedObject -Id $Device.id

        if ($backupEnvSettings) {
            foreach ($name in $backupEnvSettings.Keys) {
                if ($null -ne $name) {
                    [environment]::SetEnvironmentVariable($name, $backupEnvSettings[$name], "process")
                }
            }
        }
    }
}
