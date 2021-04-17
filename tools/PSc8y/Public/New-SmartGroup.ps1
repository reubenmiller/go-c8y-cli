# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-SmartGroup {
<#
.SYNOPSIS
Create smart group

.DESCRIPTION
Create a smart group (managed object) which groups devices by an inventory query.


.LINK
c8y smartgroups create

.EXAMPLE
PS> New-SmartGroup -Name $smartgroupName

Create smart group

.EXAMPLE
PS> New-SmartGroup -Name $smartgroupName -Data @{ myValue = @{ value1 = $true } }

Create smart group with custom properties

.EXAMPLE
PS> New-SmartGroup -Template "{ name: '$smartgroupName' }"


Create smart group using a template


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Smart group name
        [Parameter()]
        [string]
        $Name,

        # Smart group query. Should be a valid inventory query. i.e. \"name eq 'myname' and has(myFragment)\"
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Query,

        # Should the smart group be hidden from the user interface
        [Parameter()]
        [switch]
        $Invisible
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "smartgroups create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Query `
            | Group-ClientRequests `
            | c8y smartgroups create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Query `
            | Group-ClientRequests `
            | c8y smartgroups create $c8yargs
        }
        
    }

    End {}
}
