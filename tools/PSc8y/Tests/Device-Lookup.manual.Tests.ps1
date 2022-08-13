. $PSScriptRoot/imports.ps1

Describe -Name "Device lookup up manual tests" {
    BeforeEach {
        $Device = PSc8y\New-TestDevice
        $Device2 = PSc8y\New-TestDevice
        $Group = PSc8y\New-TestDeviceGroup

    }

    It "Add a device to a group using ids should only result in 1 API call (and Force=True)" {
        $output = $( $Response = PSc8y\Add-ChildAssetToDeviceGroup -Group $Group.id -Child $Device.id -Verbose -Force ) 2>&1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $output | Should -ContainRequest "GET" -Total 0
        $output | Should -ContainRequest "PUT" -Total 0
        $output | Should -ContainRequest "DELETE" -Total 0
        $output | Should -ContainRequest "POST" -Total 1
        $output | Should -ContainRequest "POST /inventory/managedObjects/$($Group.id)/childAssets" -Total 1
    }

    It "Add a device to a group using names should only cause 1 lookup per name" {
        $output = $( $Response = PSc8y\Add-ChildAssetToDeviceGroup -Group $Group.name -Child $Device.id -Verbose -Force ) 2>&1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $output | Should -ContainRequest "GET" -Total 1
        $output | Should -ContainRequest "PUT" -Total 0
        $output | Should -ContainRequest "DELETE" -Total 0
        $output | Should -ContainRequest "POST" -Total 1
        $output | Should -ContainRequest "GET /inventory/managedObjects" -Total 1
        $output | Should -ContainRequest "POST /inventory/managedObjects/$($Group.id)/childAssets" -Total 1
    }

    It "Should not do name unnecessary name lookups if the confirmation preference is set to None" {
        $ConfirmBackup = $ConfirmPreference
        $ConfirmPreference = "None"
        $output = $( $Response = PSc8y\Update-Device -Id $Device.id -NewName "mytestname" -Verbose ) 2>&1
        $ConfirmPreference = $ConfirmBackup
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $output | Should -ContainRequest "GET" -Total 0
        $output | Should -ContainRequest "PUT" -Total 1
        $output | Should -ContainRequest "DELETE" -Total 0
        $output | Should -ContainRequest "POST" -Total 0
        $output | Should -ContainRequest "PUT /inventory/managedObjects/$($Device.id)" -Total 1
    }

    It "Should not do name unnecessary name lookups if the Force parameter is used" {
        $output = $( $Response = PSc8y\Update-Device -Id $Device.id -NewName "mytestname" -Verbose -Force ) 2>&1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $output | Should -ContainRequest "GET" -Total 0
        $output | Should -ContainRequest "PUT" -Total 1
        $output | Should -ContainRequest "DELETE" -Total 0
        $output | Should -ContainRequest "POST" -Total 0
        $output | Should -ContainRequest "PUT /inventory/managedObjects/$($Device.id)" -Total 1
    }

    It "Should only use 1 api call when looking up managed object by an ID" {
        $output = $( $Response = PSc8y\Get-Device -Id $Device.id -Verbose ) 2>&1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $output | Should -ContainRequest "GET" -Total 1
        $output | Should -ContainRequest "PUT" -Total 0
        $output | Should -ContainRequest "DELETE" -Total 0
        $output | Should -ContainRequest "POST" -Total 0
        $output | Should -ContainRequest "GET /inventory/managedObjects/$($Device.id)" -Total 1
    }

    It "Accepts multiple named values in the --id parameter" {
        $output = c8y devices get --id $Device.name,$Device2.name --select id,name --output csv
        $LASTEXITCODE | Should -Be 0
        $output | Should -HaveCount 2
        $output[0] | Should -BeExactly "$($Device.id),$($($Device.name))"
        $output[1] | Should -BeExactly "$($Device2.id),$($($Device2.name))"
    }

    AfterEach {
        PSc8y\Remove-ManagedObject -Id $Device.id
        PSc8y\Remove-ManagedObject -Id $Group.id

    }
}

