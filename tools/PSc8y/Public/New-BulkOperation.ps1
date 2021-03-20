# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-BulkOperation {
<#
.SYNOPSIS
Create bulk operation

.DESCRIPTION
Create a new bulk operation

.LINK
c8y bulkoperations create

.EXAMPLE
PS> New-BulkOperation -Group $Group.id -StartDate "60s" -CreationRampSec 15 -Operation @{ c8y_Restart = @{} }

Create bulk operation for a group

.EXAMPLE
PS> Get-DeviceGroup $Group.id | New-BulkOperation -StartDate "10s" -CreationRampSec 15 -Operation @{ c8y_Restart = @{} }

Create bulk operation for a group (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Identifies the target group on which this operation should be performed.
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Group,

        # Time when operations should be created. Defaults to 300s
        [Parameter()]
        [string]
        $StartDate,

        # Delay between every operation creation. (required)
        [Parameter(Mandatory = $true)]
        [float]
        $CreationRampSec,

        # Operation prototype to send to each device in the group (required)
        [Parameter(Mandatory = $true)]
        [object]
        $Operation
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "bulkoperations create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.bulkoperation+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Group `
            | Group-ClientRequests `
            | c8y bulkoperations create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Group `
            | Group-ClientRequests `
            | c8y bulkoperations create $c8yargs
        }
        
    }

    End {}
}
