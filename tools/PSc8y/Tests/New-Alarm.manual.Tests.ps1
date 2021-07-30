. $PSScriptRoot/imports.ps1

Describe -Name "New-Alarm" {
    BeforeEach {
    }

    It "Create a new alarm for device using a template" {
        $Template = '{ type: \"c8y_TestAlarm\", time: \"2021-07-12T17:35:12Z\", text: \"Test alarm\", severity: \"MAJOR\"}'
        $Request = PSc8y\New-Alarm -Device 1234 -Template $Template -Dry -DryFormat json
        $LASTEXITCODE | Should -Be 0

        $Request | Should -Not -BeNullOrEmpty
        $Body = ($Request | ConvertFrom-Json).body
        $Body.source.id | Should -BeExactly "1234"
        $Body.type | Should -BeExactly "c8y_TestAlarm"
        $Body.text | Should -BeExactly "Test alarm"
        $Body.severity | Should -BeExactly "MAJOR"
        $Request | Should -BeLike '*"time":"2021-07-12T17:35:12Z"*'
    }

    AfterEach {
    }
}

