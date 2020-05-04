. $PSScriptRoot/imports.ps1

Describe -Name "Get-AuditRecord" {
    BeforeEach {
        $Device = New-TestDevice
        $Record = New-AuditRecord -Type "ManagedObject" -Time "0s" -Text "Managed Object updated: my_Prop: value" -Source $Device.id -Activity "Managed Object updated" -Severity "information"

    }

    It "Get an audit record by id" {
        $Response = PSc8y\Get-AuditRecord -Id $Record.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id

    }
}

