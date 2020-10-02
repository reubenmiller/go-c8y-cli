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

    It "Get all of the alarms using IncludeAll and custom include all page size" {
        $env:C8Y_SETTINGS_INCLUDEALL_PAGESIZE = "10"

        $Response = PSc8y\Get-AlarmCollection `
            -Device $Device.id `
            -IncludeAll `
            -Verbose 4> $cliOutputFile

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response | Should -HaveCount 20

        $VerboseOutput = Get-Content $cliOutputFile

        ($VerboseOutput -match "settings.includeAll.pageSize") | Should -BeLike "*settings.includeAll.pageSize: 10"

        # 2 because the first result does not have the "fetching next page"
        ($VerboseOutput -match "Fetching next page").Count | Should -BeExactly 2
    }

    It "Get all of the alarms using IncludeAll and uneven custom include size" {
        $env:C8Y_SETTINGS_INCLUDEALL_PAGESIZE = "12"

        $Response = PSc8y\Get-AlarmCollection `
            -Device $Device.id `
            -IncludeAll `
            -Verbose 4> $cliOutputFile

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response | Should -HaveCount 20

        $VerboseOutput = Get-Content $cliOutputFile

        ($VerboseOutput -match "settings.includeAll.pageSize") | Should -BeLike "*settings.includeAll.pageSize: 12"

        # 1 because only one extra fetch should be required
        # as the first has 12 results, and the second result set has less than the requested
        # page size, so it should not try to fetch another page
        ($VerboseOutput -match "Fetching next page").Count | Should -BeExactly 1
    }

    It "Using include All with WhatIf" {
        $env:C8Y_SETTINGS_INCLUDEALL_PAGESIZE = ""

        $Response = PSc8y\Get-DeviceCollection `
            -IncludeAll `
            -WhatIf `
            -Verbose 4> $cliOutputFile

        $LASTEXITCODE | Should -Be 0
        $Response | Should -BeNullOrEmpty

        $VerboseOutput = Get-Content $cliOutputFile

        ($VerboseOutput -match "settings.includeAll.pageSize") | Should -BeLike "*settings.includeAll.pageSize: 2000"
    }

    It "Set default pagesize using environment setting" {
        $env:C8Y_SETTINGS_DEFAULT_PAGESIZE = "10"

        $Response = PSc8y\Get-AlarmCollection `
            -Device $Device.id `
            -Verbose 4> $cliOutputFile

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response | Should -HaveCount 10

        # $VerboseOutput = Get-Content $cliOutputFile
        # ($VerboseOutput -match "settings.default.pageSize") | Should -BeLike "*settings.default.pageSize: 10"
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

    }
}

