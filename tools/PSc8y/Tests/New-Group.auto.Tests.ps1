. $PSScriptRoot/imports.ps1

Describe -Name "New-Group" {
    BeforeEach {
        $GroupName = "testgroup_" + [guid]::NewGuid().Guid.Substring(1,10)

    }

    It "Create a user group" {
        $Response = PSc8y\New-Group -Name "$GroupName"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-GroupByName -Name "$GroupName" | Remove-Group

    }
}

