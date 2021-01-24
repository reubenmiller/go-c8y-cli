#Requires -Version 5.0
<#
    .SYNOPSIS
        Writes messages to the information stream, optionally with
        color when written to the host.
    .DESCRIPTION
        An alternative to Write-Host which will write to the information stream
        and the host (optionally in colors specified) but will honor the
        $InformationPreference of the calling context.
        In PowerShell 5.0+ Write-Host calls through to Write-Information but
        will _always_ treats $InformationPreference as 'Continue', so the caller
        cannot use other options to the preference variable as intended.
    
    .NOTES
        Reference: https://blog.kieranties.com/2018/03/26/write-information-with-colours
#>
Function Write-InformationColored {
    [CmdletBinding()]
    param(
        # Message data
        [Parameter(Mandatory = $true)]
        [Object] $MessageData,

        # Foreground color
        [System.ConsoleColor] $ForegroundColor, # Make sure we use the current colours by default

        # Background color
        [System.ConsoleColor] $BackgroundColor,

        # Do not append a newline character
        [Switch] $NoNewline,

        # Show the information on the host by default. This will set -InformationAction to continue
        # so that the message is displayed on the console
        [switch] $ShowHost,

        # Tags to add to the information
        [string[]] $Tags
    )

    $msg = [System.Management.Automation.HostInformationMessage]@{
        Message         = $MessageData
        ForegroundColor = $ForegroundColor
        BackgroundColor = $BackgroundColor
        NoNewline       = $NoNewline.IsPresent
    }

    $options = @{
        MessageData = $msg
        Tags = $Tags
    }

    if ($ShowHost) {
        $options["InformationAction"] = "Continue"
    }

    Write-Information @options
}
