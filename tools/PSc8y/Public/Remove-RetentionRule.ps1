# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-RetentionRule {
<#
.SYNOPSIS
Delete retention rule

.DESCRIPTION
Delete an existing retention rule


.LINK
c8y retentionrules delete

.EXAMPLE
PS> Remove-RetentionRule -Id $RetentionRule.id

Delete a retention rule

.EXAMPLE
PS> Get-RetentionRule -Id $RetentionRule.id | Remove-RetentionRule

Delete a retention rule (using pipeline)


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
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "retentionrules delete"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y retentionrules delete $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y retentionrules delete $c8yargs
        }
        
    }

    End {}
}
