. $PSScriptRoot/imports.ps1

Describe -Name "Common parameters" {

    It "All commands support common c8y parameters" {
        $ExcludeCmdlets = @(
            "Get-ClientBinary",
            "Get-ClientBinaryVersion",
            "Get-CurrentTenantApplicationCollection",
            "Install-ClientBinary",
            "Invoke-BinaryProcess"
        )

        $cmdlets = Get-Command -Module PSc8y -Name "*" |
            Where-Object {
                $_.Name -match "Alarm|Event|Binary|Application|User|Group|Role|Tenant|Microservice"
            } |
            Where-Object {
                $_.Name -notmatch "Expand-|Watch-|-Test" -and $ExcludeCmdlets -notcontains $_.Name
            }

        foreach ($icmdlet in $cmdlets) {
            $icmdlet | Should -HaveParameter "OutputFile"
            $icmdlet | Should -HaveParameter "NoProxy"
            $icmdlet | Should -HaveParameter "Session"
            $icmdlet | Should -HaveParameter "TimeoutSec"
        }
    }

    It "All side-effect commands support WhatIf" {
        $ExcludeCmdlets = @(
            "Add-PowershellType",
            "New-RandomPassword",
            "New-RandomString",
            "New-Session",
            "Register-Alias",
            "Set-Session",
            "Set-ClientConsoleSetting"
        )

        $cmdlets = Get-Command -Module PSc8y -Name "*" | Where-Object {
            $_.Name -match "^(Add|Update|Remove|Set|Reset|Register|New|Enable|Disable|Approve)"
        } | Where-Object {
            $ExcludeCmdlets -notcontains $_.name
        }

        foreach ($icmdlet in $cmdlets) {
            $icmdlet | Should -HaveParameter "WhatIf"
            $icmdlet | Should -HaveParameter "Force"
        }
    }

    It "All new and update commands support templates" {
        $ExcludeCmdlets = @(
            "Add-PowershellType",
            "New-RandomPassword",
            "New-RandomString",
            "New-Session",
            "Register-Alias",
            "Set-Session",
            "Set-ClientConsoleSetting",
            "New-TestFile",
            "New-Microservice"
        )

        $cmdlets = Get-Command -Module PSc8y -Name "*" | Where-Object {
            $_.Name -match "^(Add|Update|Set|Reset|Register|New|Enable|Approve)"
        } | Where-Object {
            $ExcludeCmdlets -notcontains $_.name
        }

        foreach ($icmdlet in $cmdlets) {
            $icmdlet | Should -HaveParameter "Template"
            $icmdlet | Should -HaveParameter "TemplateVars"
        }
    }

    It "All commands with a Device parameter supports pipeline input" {
        $ExcludeCmdlets = @(
            "Add-ChildDeviceToDevice",
            "Get-ChildAssetCollection",
            "Get-ManagedObjectCollection",
            "Remove-AlarmCollection",
            "Remove-EventCollection",
            "Remove-OperationCollection"
        )
        $cmdlets = Get-Command -Module PSc8y -Name "*" -CommandType Function |
            Where-Object {
                $ExcludeCmdlets -notcontains $_.name
            } |
            Where-Object {
                $_.Parameters.Device
            }

        foreach ($icmdlet in $cmdlets) {
            $icmdlet.Parameters.Device.Attributes.ValueFromPipeline | Should -BeExactly $true -Because "$($icmdlet.Name) should support device pipes"
            $icmdlet.Parameters.Device.Attributes.ValueFromPipelineByPropertyName | Should -BeExactly $true -Because "$($icmdlet.Name) should support device pipes"
        }
    }
    

    It "All commands support Verbose" {
        $cmdlets = Get-Command -Module PSc8y -Name "*" -CommandType Function

        foreach ($icmdlet in $cmdlets) {
            $icmdlet | Should -HaveParameter "Verbose"
        }
    }

    It "Using -WhatIf should show output on the console" {
        $response = PSc8y\New-Device `
            -Name "testme" `
            -Whatif 6>&1
        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        ($Response -join "`n") | Should -BeLike "*/inventory/managedObject*"
    }
}
