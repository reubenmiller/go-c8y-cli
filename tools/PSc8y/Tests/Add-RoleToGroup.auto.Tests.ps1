. $PSScriptRoot/imports.ps1

Describe -Name "Add-RoleToGroup" {
    BeforeEach {
        $Group = New-TestGroup -Name "customGroup1"

    }

    It "Add a role to a group using wildcards" {
        $Response = PSc8y\Add-RoleToGroup -Group "customGroup1*" -Role "*ALARM_*"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Add a role to a group using wildcards (using pipeline)" {
        $Response = PSc8y\Get-RoleCollection -PageSize 100 | Where-Object Name -like "*ALARM*" | Add-RoleToGroup -Group "customGroup1*"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-Group -Id $Group.id

    }
}

