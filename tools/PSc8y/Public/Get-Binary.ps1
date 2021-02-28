# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-Binary {
<#
.SYNOPSIS
Get binary

.DESCRIPTION
Download a binary stored in Cumulocity and display it on the console. For non text based binaries or if the output should be saved to file, the output parameter should be used to write the file directly to a local file.


.LINK
c8y binaries get

.EXAMPLE
PS> Get-Binary -Id $Binary.id

Get a binary and display the contents on the console

.EXAMPLE
PS> Get-Binary -Id $Binary.id -OutputFile ./download-binary1.txt

Get a binary and save it to a file


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Inventory binary id (required)
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "binaries get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "*/*"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y binaries get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y binaries get $c8yargs
        }
        
    }

    End {}
}
