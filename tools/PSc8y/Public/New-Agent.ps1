# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Agent {
<#
.SYNOPSIS
Create agent

.DESCRIPTION
Create an agent managed object. An agent is a special device managed object with both the
c8y_IsDevice and com_cumulocity_model_Agent fragments.


.LINK
c8y agents create

.EXAMPLE
PS> New-Agent -Name $AgentName

Create agent

.EXAMPLE
PS> New-Agent -Name $AgentName -Data @{ myValue = @{ value1 = $true } }

Create agent with custom properties


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Agent name
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Name,

        # Agent type
        [Parameter()]
        [string]
        $Type
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "agents create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.customAgent+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

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
        
    }

    End {}
}
