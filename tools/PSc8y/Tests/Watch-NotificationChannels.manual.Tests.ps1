. $PSScriptRoot/imports.ps1

Describe -Name "Watch-NotificationChannels" {
    BeforeEach {
        $Device = New-TestAgent
        Start-Sleep -Seconds 2

        # Create background task which creates measurements
        $importpath = (Resolve-Path "$PSScriptRoot/imports.ps1").ProviderPath
        $JobArgs = @(
            $importpath,
            $env:C8Y_SESSION,
            $Device.id
        )

        # Create measurements
        $Job = Start-Job -Name "watch-measurements-data" -Debug -ArgumentList $JobArgs -ScriptBlock {
            . $args[0]
            Start-Sleep -Seconds 2
            $env:C8Y_SESSION = $args[1]
            $DeviceID = $args[2]
            @(1..10) | ForEach-Object {
                New-TestMeasurement -Device $DeviceID -Force
                Start-Sleep -Milliseconds 1000
            }
        }

        # Create operations
        $Job2 = Start-Job -Name "watch-operation-data" -Debug -ArgumentList $JobArgs -ScriptBlock {
            . $args[0]
            Start-Sleep -Seconds 2
            $env:C8Y_SESSION = $args[1]
            $DeviceID = $args[2]
            @(1..10) | ForEach-Object {
                New-TestOperation -Device $DeviceID -Force
                Start-Sleep -Milliseconds 1000
            }
        }
    }

    It "Watch all notifications for a time period" {
        $StartTime = Get-Date
        [array] $Response = PSc8y\Watch-NotificationChannels -Device $Device.id -DurationSec 15
        $LASTEXITCODE | Should -Be 0
        $Response.Count | Should -BeGreaterOrEqual 0
        $Duration = (Get-Date) - $StartTime
        $Duration.TotalSeconds | Should -BeGreaterOrEqual 15
    }

    It "Watch a device for a number of notifications" {
        $Response = PSc8y\Watch-NotificationChannels -Device $Device.id -Count 2
        $LASTEXITCODE | Should -Be 0
        $Response | Should -HaveCount 2
    }

    It "Watch notifications for all devices and stop after receiving x messages" {
        $Response = PSc8y\Watch-NotificationChannels -Count 2
        $LASTEXITCODE | Should -Be 0
        $Response | Should -HaveCount 2
    }

    AfterEach {
        Stop-Job -Job $Job, $Job2
        Remove-Job -Id $Job.Id
        Remove-Job -Id $Job2.Id
        PSc8y\Remove-ManagedObject -Id $Device.id

    }
}
