﻿. $PSScriptRoot/imports.ps1

Describe -Name "Watch-Alarm" {
    BeforeEach {
        $Device = New-TestDevice
        Start-Sleep -Seconds 2

        # Create background task which creates alarms
        $importpath = (Resolve-Path "$PSScriptRoot/imports.ps1").ProviderPath
        $JobArgs = @(
            $importpath,
            $env:C8Y_SESSION,
            $Device.id
        )
        $Job = Start-Job -Name "watch-alarm-data" -Debug -ArgumentList $JobArgs -ScriptBlock {
            . $args[0]
            Start-Sleep -Seconds 2
            $env:C8Y_SESSION = $args[1]
            $DeviceID = $args[2]
            @(1..10) | ForEach-Object {
                New-TestAlarm -Device $DeviceID -Force
                Start-Sleep -Milliseconds 1000
            }
        }
    }

    It "Watch alarms for a time period" {
        $StartTime = Get-Date

        $FirstUpdate = $null
        $LastUpdate = $null

        [array] $Response = PSc8y\Watch-Alarm -Device $Device.id -DurationSec 10 | ForEach-Object {
            $now = Get-Date
            if ($null -eq $FirstUpdate) {
                $FirstUpdate = $now
            }
            $LastUpdate = $now
            $_
        }
        $LASTEXITCODE | Should -Be 0
        $Response.Count | Should -BeGreaterOrEqual 0
        $Duration = (Get-Date) - $StartTime
        $Duration.TotalSeconds | Should -BeGreaterOrEqual 10

        ($LastUpdate - $FirstUpdate).TotalSeconds |
            Should -BeGreaterThan 2 -Because "Values should be sent to pipeline as soon as they arrive"
    }

    It "Watch alarms for a number of alarms" {
        $Response = PSc8y\Watch-Alarm -Device $Device.id -Count 2
        $LASTEXITCODE | Should -Be 0
        $Response | Should -HaveCount 2
    }

    It "Watch alarms for all devices and stop after receiving x messages" {
        $Response = PSc8y\Watch-Alarm -Count 2
        $LASTEXITCODE | Should -Be 0
        $Response | Should -HaveCount 2
    }

    AfterEach {
        Stop-Job -Job $Job
        Remove-Job -Id $Job.Id
        PSc8y\Remove-ManagedObject -Id $Device.id

    }
}
