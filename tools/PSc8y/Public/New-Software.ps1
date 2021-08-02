# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-Software {
<#
.SYNOPSIS
Create software package

.DESCRIPTION
Create a new software package (managedObject)

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/software_create

.EXAMPLE
PS> New-ManagedObject -Name "python3-requests" -Description "python requests library" -Data @{$type=@{}}

Create a managed object


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # name
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Name,

        # Description of the software package
        [Parameter()]
        [string]
        $Description,

        # Device type filter. Only allow software to be applied to devices of this type
        [Parameter()]
        [string]
        $DeviceType
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "software create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.inventory+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Name `
            | Group-ClientRequests `
            | c8y software create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Name `
            | Group-ClientRequests `
            | c8y software create $c8yargs
        }
        
    }

    End {}
}
