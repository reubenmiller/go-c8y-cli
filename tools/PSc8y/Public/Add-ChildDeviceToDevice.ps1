# Code generated from specification version 1.0.0: DO NOT EDIT
Function Add-ChildDeviceToDevice {
<#
.SYNOPSIS
Assign child device

.DESCRIPTION
Create a child device reference

.LINK
c8y devices assignChild

.EXAMPLE
PS> Add-ChildDeviceToDevice -Device $Device.id -NewChild $ChildDevice.id

Assign a device as a child device to an existing device

.EXAMPLE
PS> Get-ManagedObject -Id $ChildDevice.id | Add-ChildDeviceToDevice -Device $Device.id

Assign a device as a child device to an existing device (using pipeline)


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Device. (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Device,

        # New child device (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $NewChild
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "devices assignChild"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObjectReference+json"
            ItemType = "application/vnd.com.nsn.cumulocity.managedObject+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $NewChild `
            | Group-ClientRequests `
            | c8y devices assignChild $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $NewChild `
            | Group-ClientRequests `
            | c8y devices assignChild $c8yargs
        }
        
    }

    End {}
}
