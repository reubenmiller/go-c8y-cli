. $PSScriptRoot/imports.ps1

Describe -Name "Remove-AuditRecordCollection" {
    BeforeEach {
        $Device = New-TestDevice
        $Record = New-AuditRecord -Type "ManagedObject" -Time "0s" -Text "Managed Object updated: my_Prop: value" -Source $Device.id -Activity "Managed Object updated" -Severity "information"

    }

    It "Delete audit records from a device" {
        $Response = PSc8y\Remove-AuditRecordCollection -Source $Device.id
        $LASTEXITCODE | Should -Be 0
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id

    }
}

