. $PSScriptRoot/imports.ps1

Describe -Name "Get-RoleReferenceCollectionFromGroup" {
    BeforeEach {
        $Group = Get-UserGroupByName -Name "business"

    }

    It "Get a list of role references for a user group" {
        $Response = PSc8y\Get-RoleReferenceCollectionFromGroup -Group $Group.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {

    }
}

