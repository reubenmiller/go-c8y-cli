. $PSScriptRoot/imports.ps1

Describe -Name "Watch-Notifcation" {
    BeforeEach {
        $Device = New-TestDevice

        # Create background task which creates measurements
        $importpath = (Resolve-Path "$PSScriptRoot/imports.ps1").ProviderPath
        $JobArgs = @(
            $importpath,
            $env:C8Y_SESSION,
            $Device.id
        )
        $Job = Start-Job -Name "watch-notifications-measurements-data" -Debug -ArgumentList $JobArgs -ScriptBlock {
            . $args[0]
            $env:C8Y_SESSION = $args[1]
            $DeviceID = $args[2]
            @(1..10) | ForEach-Object {
                New-TestMeasurement -Device $DeviceID -Force
                Start-Sleep -Milliseconds 500
            }
        }
    }

    It "Watch measurements for a time period" {
        $StartTime = Get-Date
        [array] $Response = PSc8y\Watch-Notification -Channel ("/measurements/" + $Device.id) -DurationSec 10
        $LASTEXITCODE | Should -Be 0
        $Response.Count | Should -BeGreaterOrEqual 0
        $Duration = (Get-Date) - $StartTime
        $Duration.TotalSeconds | Should -BeGreaterOrEqual 10
    }

    It "Watch measurements for a number of measurements" {
        $Response = PSc8y\Watch-Notification -Channel ("/measurements/" + $Device.id) -Count 2
        $LASTEXITCODE | Should -Be 0
        $Response | Should -HaveCount 2
    }

    It "Watch measurements for all devices and stop after receiving x messages" {
        $Response = PSc8y\Watch-Notification -Channel ("/measurements/" + $Device.id) -Count 2
        $LASTEXITCODE | Should -Be 0
        $Response | Should -HaveCount 2
    }

    AfterEach {
        Stop-Job -Job $Job
        Remove-Job -Id $Job.Id
        PSc8y\Remove-ManagedObject -Id $Device.id

    }
}
