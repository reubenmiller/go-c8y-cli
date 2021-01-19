. $PSScriptRoot/imports.ps1

Describe -Name "Device lookup up manual tests" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice
        $Group = PSc8y\New-TestDeviceGroup

    }

    It "Add a device to a group using ids should only result in 1 API call (and Force=True)" {
        $VerboseMessages = $( $Response = PSc8y\Add-DeviceToGroup -Group $Group.id -NewChildDevice $Device.id -Verbose -Force ) 4>&1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        [array] $APICalls = $VerboseMessages -like "*Sending request*"
        $APICalls | Should -HaveCount 1
    }

    It "Add a device to a group using names should only cause 1 lookup per name" {
        $VerboseMessages = $( $Response = PSc8y\Add-DeviceToGroup -Group $Group.name -NewChildDevice $Device.name -Verbose -Force ) 4>&1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        [array] $APICalls = $VerboseMessages -like "*Sending request*"
        $APICalls | Should -HaveCount 3 -Because "2 GETs to get device ids and 1 POST to create the child link"
    }

    It "Should not do name unnecessary name lookups if the confirmation preference is set to None" {
        $ConfirmBackup = $ConfirmPreference
        $ConfirmPreference = "None"
        $VerboseMessages = $( $Response = PSc8y\Update-Device -Id $Device.id -NewName "mytestname" -Verbose ) 4>&1
        $ConfirmPreference = $ConfirmBackup
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        [array] $APICalls = $VerboseMessages -like "*Sending request*"
        $APICalls | Should -HaveCount 1
    }

    It "Should not do name unnecessary name lookups if the Force parameter is used" {
        $VerboseMessages = $( $Response = PSc8y\Update-Device -Id $Device.id -NewName "mytestname" -Verbose -Force ) 4>&1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        [array] $APICalls = $VerboseMessages -like "*Sending request*"
        $APICalls | Should -HaveCount 1
    }

    It "Should only use 1 api call when looking up managed object by an ID" {
        $VerboseMessages = $( $Response = PSc8y\Get-Device -Id $Device.id -Verbose ) 4>&1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        [array] $APICalls = $VerboseMessages -like "*Sending request*"
        $APICalls | Should -HaveCount 1
    }

    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id
        PSc8y\Remove-ManagedObject -Id $Group.id

    }
}

