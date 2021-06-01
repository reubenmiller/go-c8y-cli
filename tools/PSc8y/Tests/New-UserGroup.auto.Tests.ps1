. $PSScriptRoot/imports.ps1

Describe -Name "New-UserGroup" {
    BeforeEach {
        $GroupName = "testgroup_" + [guid]::NewGuid().Guid.Substring(1,10)

    }

    It "Create a user group" {
        $Response = PSc8y\New-UserGroup -Name "$GroupName"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-UserGroupByName -Name "$GroupName" | Remove-UserGroup

    }
}

