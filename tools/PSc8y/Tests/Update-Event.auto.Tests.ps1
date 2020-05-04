. $PSScriptRoot/imports.ps1

Describe -Name "Update-Event" {
    BeforeEach {
        $Device = New-TestDevice
        $Event = New-TestEvent -Device $Device.id

    }

    It "Update the text field of an existing event" {
        $Response = PSc8y\Update-Event -Id $Event.id -Text "example text 1"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update custom properties of an existing event" {
        $Response = PSc8y\Update-Event -Id $Event.id -Data @{ my_event = @{ active = $true } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Update custom properties of an existing event (using pipeline)" {
        $Response = PSc8y\Get-Event -Id $Event.id | Update-Event -Data @{ my_event = @{ active = $true } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id

    }
}

