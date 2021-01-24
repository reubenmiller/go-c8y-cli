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
        $Response = PSc8y\New-ManagedObject -Name "testMO" -Template $template -TemplateVars "subtype=customName"

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

    It "Handles large jsonnet templates" {

        $Template = @"
local timestamp = var("timestamp");
local eventCount = var("eventCount", 5);
    
local CreateEvent(major='1',minor='1',extid='123456789',timestamp='2021-01-20T12:21:35.3735228+01:00') = {
    "params":{
    "c8y_UpdateEvent": {
        "major": major,
        "minor": minor,
        },
    },
    "time": timestamp,
    "source": {
        "externalId": extid,
        "type": "my_SpecialType",
        "type1": "my_SpecialType",
        "type2": "my_SpecialType",
        "type3": "my_SpecialType",
        "type4": "my_SpecialType",
        "type5": "my_SpecialType",
        "type6": "my_SpecialType",
    }, 
    "type": "c8y_UpdateEvent",
    "text": "Sub component update"
};
        
{
    "type": "c8y_Update",
    "text": "Component Update",
    "c8y_Update": {
        "subComponentId":"a5d1c5d4",
        "events":[
            CreateEvent() for i in std.range(1,eventCount)
        ],
        "measurements": []
    }
}
"@.TrimEnd()

        $TemplateFile = New-TemporaryFile
        $template | Out-File $TemplateFile

        $TestDevice = New-TestDevice
        $TestDevice | Should -Not -BeNullOrEmpty
        $null = $items.Add($TestDevice.id)

        $Response = New-Event `
            -Device $TestDevice.id `
            -Template $TemplateFile `
            -TemplateVars "timestamp=$(Format-Date),eventCount=5" `
            -WhatIf *>&1

        $Response | Should -Not -BeNullOrEmpty
        $LASTEXITCODE | Should -BeExactly 0
    }

    AfterEach {
        foreach ($item in $items) {
            if ($item) {
                PSc8y\Remove-ManagedObject -Id $item
            }
        }
    }
}
