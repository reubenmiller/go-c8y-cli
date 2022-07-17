# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-EventBinary {
<#
.SYNOPSIS
Get event binary

.DESCRIPTION
Get the binary associated with an event

When downloading a binary it is useful to use the outputFileRaw global parameter and to use one of the following references:

* {filename} - Filename found in the Content-Disposition response header
* {id} - An id like value found in the request path (/event/events/12345/binaries => 12345)
* {basename} - The last path section of the request path (/some/nested/url/withafilename.json => withafilename.json)


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/events_downloadBinary

.EXAMPLE
PS> Get-EventBinary -Id $Event.id -OutputFileRaw ./eventbinary.txt

Download a binary related to an event


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Event id (required)
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "events downloadBinary"
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
            | c8y events downloadBinary $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y events downloadBinary $c8yargs
        }
        
    }

    End {}
}
