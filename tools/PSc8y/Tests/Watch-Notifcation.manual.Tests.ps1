. $PSScriptRoot/imports.ps1

Describe -Name "Watch-Notifcation" -Skip {
    BeforeEach {
        $Device = New-TestDevice
        Start-Sleep -Seconds 5

        # Create background task which creates measurements
        $importpath = (Resolve-Path "$PSScriptRoot/imports.ps1").ProviderPath
        $JobArgs = @(
            $importpath,
            $env:C8Y_SESSION,
            $Device.id
        )
        $Job = Start-Job -Name "watch-notifications-measurements-data" -Debug -ArgumentList $JobArgs -ScriptBlock {
            $env:C8Y_SESSION = $args[1]
            . $args[0]
            Start-Sleep -Seconds 2
            $DeviceID = $args[2]
            @(1..60) | ForEach-Object {
                New-Measurement -Template "test.measurement.jsonnet" -Device $DeviceID -Force
                Start-Sleep -Milliseconds 1000
            }
        }
    }

    It "Watch measurements for a time period" {
        $StartTime = Get-Date

        [array] $Response = PSc8y\Watch-Notification -Channel ("/measurements/" + $Device.id) -Duration "60s" | ForEach-Object {
            $_ | Add-Member -MemberType NoteProperty -Name "PSc8yTimestamp" -Value (Get-Date) -PassThru
        }
        $LASTEXITCODE | Should -Be 0
        $Response.Count | Should -BeGreaterOrEqual 2
        $Duration = (Get-Date) - $StartTime
        $Duration.TotalSeconds | Should -BeGreaterOrEqual 10

        ($Response[-1].PSc8yTimestamp - $Response[0].PSc8yTimestamp).TotalSeconds |
            Should -BeGreaterThan 2 -Because "Values should be sent to pipeline as soon as they arrive"
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
