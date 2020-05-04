. $PSScriptRoot/imports.ps1

Describe -Name "New-User" {
    BeforeEach {
        $Username = "testuser_" + [guid]::NewGuid().Guid.Substring(1,10)
        $NewPassword = [guid]::NewGuid().Guid.Substring(1,10)

    }

    It "Create a user" {
        $Response = PSc8y\New-user -Username "$Username" -Password "$NewPassword"
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-UserByName -Name "$Username" | Remove-User

    }
}

