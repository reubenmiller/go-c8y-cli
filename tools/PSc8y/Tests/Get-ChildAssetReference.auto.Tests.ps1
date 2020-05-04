. $PSScriptRoot/imports.ps1

Describe -Name "Get-ChildAssetReference" {
    BeforeEach {
        $Agent = New-TestDevice -AsAgent
        $Device = New-TestDevice
        $Ref = New-ChildAssetReference -Group $Agent.id -NewChildDevice $Device.id

    }

    It "Get an existing child asset reference" {
        $Response = PSc8y\Get-ChildAssetReference -Asset $Agent.id -Reference $Ref.id
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
    }


    AfterEach {
        Remove-ManagedObject -Id $Device.id
        Remove-ManagedObject -Id $Agent.id

    }
}

