
. $PSScriptRoot/imports.ps1

Describe -Name "Expand-Source" {
    BeforeEach {
        # Must be an agent so it accepts operations
        $Device = PSc8y\New-TestAgent
    }

    It "Expand source id from a list of alarms" {
        # Note, these alarms will be dededuplicated because they have the same
        # alarm type. The count will be set to 2
        $Alarm1 = PSc8y\New-TestAlarm -Device $Device.id
        $Alarm2 = PSc8y\New-TestAlarm -Device $Device.id

        $Response = Get-AlarmCollection -Device $Device.id | PSc8y\Expand-Source
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response | Should -HaveCount 1
    }

    It "Expand source id from a list of events" {
        $Event1 = PSc8y\New-TestEvent -Device $Device.id
        $Event2 = PSc8y\New-TestEvent -Device $Device.id

        $Response = Get-EventCollection -Device $Device.id | PSc8y\Expand-Source
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response | Should -HaveCount 2
    }

    It "Expand source id from a list of measurements" {
        $Measurement1 = PSc8y\New-Measurement -Template "test.measurement.jsonnet" -Device $Device.id
        $Measurement2 = PSc8y\New-Measurement -Template "test.measurement.jsonnet" -Device $Device.id

        $Response = Get-MeasurementCollection -Device $Device.id | PSc8y\Expand-Source
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response | Should -HaveCount 2
    }

    It "Expand source id from a list of operations" {
        $Operation1 = PSc8y\New-TestOperation -Device $Device.id
        $Operation2 = PSc8y\New-TestOperation -Device $Device.id

        $Response = Get-OperationCollection -Device $Device.id | PSc8y\Expand-Source
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response | Should -HaveCount 2
    }

    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id
    }
}
