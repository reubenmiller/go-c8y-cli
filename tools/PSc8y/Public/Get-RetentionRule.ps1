# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-RetentionRule {
<#
.SYNOPSIS
Get retention rule

.DESCRIPTION
Get an existing retention by id


.LINK
c8y retentionRules get

.EXAMPLE
PS> Get-RetentionRule -Id $RetentionRule.id

Get a retention rule


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Retention rule id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "retentionRules get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.retentionRule+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y retentionRules get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y retentionRules get $c8yargs
        }
        
    }

    End {}
}
