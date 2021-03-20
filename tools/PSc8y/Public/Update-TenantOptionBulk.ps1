# Code generated from specification version 1.0.0: DO NOT EDIT
Function Update-TenantOptionBulk {
<#
.SYNOPSIS
Update multiple tenant options

.DESCRIPTION
Update multiple tenant options in provided category

.LINK
c8y tenantoptions updateBulk

.EXAMPLE
PS> Update-TenantOptionBulk -Category "c8y_cli_tests" -Data @{ $option5 = 0; $option6 = 1 }

Update multiple tenant options


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Tenant Option category (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Category
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Update", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "tenantoptions updateBulk"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.option+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Category `
            | Group-ClientRequests `
            | c8y tenantoptions updateBulk $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Category `
            | Group-ClientRequests `
            | c8y tenantoptions updateBulk $c8yargs
        }
        
    }

    End {}
}
