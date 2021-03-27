# Code generated from specification version 1.0.0: DO NOT EDIT
Function Add-ChildGroupToGroup {
<#
.SYNOPSIS
Assign child group

.DESCRIPTION
Assigns a group to a group. The group will be a childAsset of the group

.LINK
c8y devicegroups assignGroup

.EXAMPLE
PS> Add-ChildGroupToGroup -Group $Group.id -NewChildGroup $ChildGroup1.id

Add a group to a group as a child

.EXAMPLE
PS> Get-DeviceGroup $SubGroup1.name, $SubGroup2.name | Add-ChildGroupToGroup -Group $CustomGroup.id

Add multiple devices to a group. Alternatively `Get-DeviceCollection` can be used
to filter for a collection of devices and assign the results to a single group.



#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Group (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Group,

        # New child group to be added to the group as an child asset (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $NewChildGroup
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devicegroups assignGroup"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReference+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $NewChildGroup `
            | Group-ClientRequests `
            | c8y devicegroups assignGroup $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $NewChildGroup `
            | Group-ClientRequests `
            | c8y devicegroups assignGroup $c8yargs
        }
        
    }

    End {}
}
