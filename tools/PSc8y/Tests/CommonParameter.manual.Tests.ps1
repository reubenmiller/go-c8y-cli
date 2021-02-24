. $PSScriptRoot/imports.ps1

Describe -Name "Common parameters" {

    It "All commands support common c8y parameters" {
        $ExcludeCmdlets = @(
            "Get-ClientBinary",
            "Get-ClientBinaryVersion",
            "Get-CurrentTenantApplicationCollection",
            "Group-ClientRequests",
            "ConvertFrom-ClientOutput",
            "Install-ClientBinary"
        )

        $cmdlets = Get-Command -Module PSc8y -Name "*" |
            Where-Object {
                $_.Name -match "Alarm|Event|Binary|Application|User|Group|Role|Tenant|Microservice"
            } |
            Where-Object {
                $_.Name -notmatch "Expand-|Watch-|-Test" -and $ExcludeCmdlets -notcontains $_.Name
            }

        foreach ($icmdlet in $cmdlets) {
            $resolvedCommand = $icmdlet
            if ($icmdlet.CommandType -eq "Alias") {
                $resolvedCommand = $icmdlet.ResolvedCommand
            }
            $resolvedCommand | Should -HaveParameter "OutputFile"
            $resolvedCommand | Should -HaveParameter "NoProxy"
            $resolvedCommand | Should -HaveParameter "Session"
            $resolvedCommand | Should -HaveParameter "Timeout"
        }
    }

    It "All side-effect commands support WhatIf" {
        $ExcludeCmdlets = @(
            "Add-PowershellType",
            "Add-ClientResponseType",
            "New-RandomPassword",
            "New-RandomString",
            "New-Session",
            "Register-Alias",
            "Register-ClientArgumentCompleter",
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
            "Add-ClientResponseType",
            "New-RandomPassword",
            "New-RandomString",
            "New-Session",
            "Register-Alias",
            "Register-ClientArgumentCompleter",
            "Set-Session",
            "Set-ClientConsoleSetting",
            "New-TestFile",
            "New-Microservice",
            "New-ServiceUser"
        )

        $cmdlets = Get-Command -Module PSc8y -Name "*" | Where-Object {
            $_.Name -match "^(Add|Update|Set|Reset|Register|New|Enable|Approve|Invoke-ClientRequest)"
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
            "Remove-OperationCollection",
            "Remove-ChildDeviceFromDevice"
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
        $output = PSc8y\New-Device `
            -Name "testme" `
            -WhatIf 2>&1
        $LASTEXITCODE | Should -Be 0
        $output | Should -Not -BeNullOrEmpty
        $output | Out-String | Should -BeLike "*/inventory/managedObject*"
    }
}
