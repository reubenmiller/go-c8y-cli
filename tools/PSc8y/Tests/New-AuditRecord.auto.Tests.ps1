. $PSScriptRoot/imports.ps1

Describe -Name "New-AuditRecord" {
    BeforeEach {
        $Device = New-TestDevice

    }

    It "Create an audit record for a custom managed object update" {
        $Response = PSc8y\New-AuditRecord -Type "ManagedObject" -Time "0s" -Text "Managed Object updated: my_Prop: value" -Source $Device.id -Activity "Managed Object updated" -Severity "information"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id

    }
}

