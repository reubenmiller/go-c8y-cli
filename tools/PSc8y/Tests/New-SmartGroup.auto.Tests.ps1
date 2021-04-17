. $PSScriptRoot/imports.ps1

Describe -Name "New-SmartGroup" {
    BeforeEach {
        $smartgroupName = PSc8y\New-RandomString -Prefix "mySmartGroup"

    }

    It "Create smart group" {
        $Response = PSc8y\New-SmartGroup -Name $smartgroupName
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create smart group with custom properties" {
        $Response = PSc8y\New-SmartGroup -Name $smartgroupName -Data @{ myValue = @{ value1 = $true } }
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }

    It "Create smart group using a template" {
        $Response = PSc8y\New-SmartGroup -Template "{ name: '$smartgroupName' }"

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-SmartGroup -Id $smartgroupName

    }
}

