# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-FirmwareVersionBinary {
<#
.SYNOPSIS
Download firmware version binary

.DESCRIPTION
Download a binary stored in Cumulocity and display it on the console. For non text based binaries or if the output should be saved to file, the output parameter should be used to write the file directly to a local file.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/firmware_versions_download

.EXAMPLE
PS> Get-Binary -Id $Binary.id

Get a binary and display the contents on the console

.EXAMPLE
PS> Get-Binary -Id $Binary.id -OutputFileRaw ./download-binary1.txt

Get a binary and save it to a file


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Firmware url (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [string[]]
        $Url
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "firmware versions download"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "*/*"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Url `
            | Group-ClientRequests `
            | c8y firmware versions download $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Url `
            | Group-ClientRequests `
            | c8y firmware versions download $c8yargs
        }
        
    }

    End {}
}
