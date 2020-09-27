. $PSScriptRoot/imports.ps1

Describe -Name "Add-RoleToGroup" {
    BeforeEach {
        $Group = New-TestGroup -Name "customGroup1"
        $NamePattern = $Group.name.Substring(0, $Group.name.length - 2)

    }

    It "Add a role to a group using wildcards" {
        $Response = PSc8y\Add-RoleToGroup -Group "${NamePattern}*" -Role "*ALARM_*"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Add a role to a group using wildcards (using pipeline)" {
        $Response = PSc8y\Get-RoleCollection -PageSize 100 | Where-Object Name -like "*ALARM*" | Add-RoleToGroup -Group "${NamePattern}*"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        PSc8y\Remove-Group -Id $Group.id

    }
}

