Function New-TestSmartGroup {
    <# 
    .SYNOPSIS
    Create a new test smart group in Cumulocity
    
    .DESCRIPTION
    Create a new test smart group with a randomized name. Useful when performing mockups or prototyping.
        
    .EXAMPLE
    New-TestSmartGroup
    
    Create a test smart group
    
    .EXAMPLE
    1..10 | Foreach-Object { New-TestSmartGroup -Force }
    
    Create 10 test smart groups all with unique names
    
    #>
        [cmdletbinding()]
        Param(
            # Device name prefix which is added before the randomized string
            [Parameter(
                Mandatory = $false,
                ValueFromPipeline = $true,
                ValueFromPipelineByPropertyName = $true,
                Position = 0
            )]
            [string] $Name
        )
        DynamicParam {
            Get-ClientCommonParameters -Type "Create", "TemplateVars"
        }
    
        Begin {
            $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "smartgroups create"
            $ClientOptions = Get-ClientOutputOption $PSBoundParameters
            $TypeOptions = @{
                Type = "application/vnd.com.nsn.cumulocity.inventory+json"
                ItemType = ""
                BoundParameters = $PSBoundParameters
            }
            $Template = ""
            if (-Not $Template) {
                $Template = (Join-Path $script:Templates "test.smartgroup.jsonnet")
            }
            [void] $c8yargs.AddRange(@(
                "--template",
                $Template
            ))
        }
    
        Process {
            if ($ClientOptions.ConvertToPS) {
                $Name `
                | Group-ClientRequests `
                | c8y smartgroups create $c8yargs `
                | ConvertFrom-ClientOutput @TypeOptions
            }
            else {
                $Name `
                | Group-ClientRequests `
                | c8y smartgroups create $c8yargs
            }
        }
    }
    