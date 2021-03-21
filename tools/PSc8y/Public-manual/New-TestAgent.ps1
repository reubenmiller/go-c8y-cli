Function New-TestAgent {
    <# 
    .SYNOPSIS
    Create test agent
    
    .DESCRIPTION
    Create a new test agent with a randomized name. Useful when performing mockups or prototyping.
    
    .EXAMPLE
    New-TestAgent
    
    Create a test agent
    
    .EXAMPLE
    1..10 | Foreach-Object { New-TestAgent -Force }
    
    Create 10 test agents all with unique names
    
    #>
        [cmdletbinding()]
        Param(
            # Agent name prefix which is added before the randomized string
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
            $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "agents create"
            $ClientOptions = Get-ClientOutputOption $PSBoundParameters
            $TypeOptions = @{
                Type = "application/vnd.com.nsn.cumulocity.customAgent+json"
                ItemType = ""
                BoundParameters = $PSBoundParameters
            }
            [void] $c8yargs.AddRange(@(
                "--template",
                "test.agent.jsonnet"
            ))
        }
    
        Process {
            if ($Name) {
                if ($ClientOptions.ConvertToPS) {
                    $Name `
                    | Group-ClientRequests `
                    | c8y agents create $c8yargs `
                    | ConvertFrom-ClientOutput @TypeOptions
                }
                else {
                    $Name `
                    | Group-ClientRequests `
                    | c8y agents create $c8yargs
                }
            } else {
                if ($ClientOptions.ConvertToPS) {
                    c8y agents create $c8yargs `
                    | ConvertFrom-ClientOutput @TypeOptions
                }
                else {
                    c8y agents create $c8yargs
                }
            }
        }
    }
    