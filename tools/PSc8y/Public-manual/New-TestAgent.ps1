Function New-TestAgent {
<# 
.SYNOPSIS
Create a new test agent representation in Cumulocity

.DESCRIPTION
Create a new test agent with a randomized name. Useful when performing mockups or prototyping.

The agent will have both the `c8y_IsDevice` and `com_cumulocity_model_Agent` fragments set.

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
        [string] $Name = "testagent",

        # Cumulocity processing mode
        [Parameter()]
        [AllowNull()]
        [AllowEmptyString()]
        [ValidateSet("PERSISTENT", "QUIESCENT", "TRANSIENT", "CEP")]
        [string]
        $ProcessingMode,

        # Template (jsonnet) file to use to create the request body.
        [Parameter()]
        [string]
        $Template,

        # Variables to be used when evaluating the Template. Accepts json or json shorthand, i.e. "name=peter"
        [Parameter()]
        [string]
        $TemplateVars,

        # Don't prompt for confirmation
        [switch] $Force
    )
    Process {
        $Data = @{
            c8y_IsDevice = @{}
            com_cumulocity_model_Agent = @{}
        }

        $AgentName = New-RandomString -Prefix "${Name}_"
        $TestAgent = PSc8y\New-ManagedObject `
            -Name $AgentName `
            -Data $Data `
            -ProcessingMode:$ProcessingMode `
            -Template:$Template `
            -TemplateVars:$TemplateVars `
            -Force:$Force

        $TestAgent
    }
}
