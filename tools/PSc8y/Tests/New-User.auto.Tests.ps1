. $PSScriptRoot/imports.ps1

Describe -Name "New-User" {
    BeforeEach {
        $Username = "testuser_" + [guid]::NewGuid().Guid.Substring(1,10)
        $NewPassword = New-RandomPassword

    }

    It "Create a user and force user to change their password when logging in" {
        $Response = PSc8y\New-user -Username "$Username" -Email "testuser@no-reply.dummy.com" -Password "$NewPassword" -ShouldResetPassword
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Get-UserByName -Name "$Username" | Remove-User

    }
}

