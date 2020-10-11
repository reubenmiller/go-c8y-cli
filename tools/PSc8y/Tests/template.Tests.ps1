. $PSScriptRoot/imports.ps1

Describe -Name "New-ManagedObject Templates" {
    BeforeEach {
        $items = New-Object System.Collections.ArrayList
    }

    It "Create a managed object using templates" {
        $template = @"
{
    type: var("type"),
    dummyFragment: {
        value: rand.int,
    },
}
"@
        $TemplateFile = New-TemporaryFile
        $template | Out-File $TemplateFile
        $Response = PSc8y\New-ManagedObject -Name "testMO" -Template $TemplateFile

        if ($Response.id) {
            $null = $items.Add($Response.id)
        }

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty
        $Response.dummyFragment.value | Should -Not -BeNullOrEmpty
        $Response.type | Should -BeExactly ""
        $Response.name | Should -BeExactly "testMO"
    }

    It "Creates a managed objects using template variables" {
        $template = @"
{
    type: var('type', 'defaultValue'),
    subtype: var('subtype'),
    values: {
        bool: rand.bool,
        int: rand.int,
        int2: rand.int2,
        float: rand.float,
        float2: rand.float2,
        float3: rand.float3,
        float4: rand.float4,
    },
}
"@
        $Response = PSc8y\New-ManagedObject -Name "testMO" -Template $template -TemplateVar "subtype=customName"

        if ($Response.id) {
            $null = $items.Add($Response.id)
        }

        $LASTEXITCODE | Should -Be 0
        $Response | Should -Not -BeNullOrEmpty

        $Response.name | Should -BeExactly "testMO"
        $Response.type | Should -BeExactly "defaultValue"
        $Response.subtype | Should -BeExactly "customName"
        $Response.values.bool | Should -Not -BeNullOrEmpty
        $Response.values.int | Should -Not -BeNullOrEmpty
        $Response.values.int2 | Should -Not -BeNullOrEmpty
        $Response.values.float | Should -Not -BeNullOrEmpty
        $Response.values.float2 | Should -Not -BeNullOrEmpty
        $Response.values.float3 | Should -Not -BeNullOrEmpty
        $Response.values.float4 | Should -Not -BeNullOrEmpty
    }

    AfterEach {
        foreach ($item in $items) {
            if ($item) {
                PSc8y\Remove-ManagedObject -Id $item
            }
        }
    }
}
