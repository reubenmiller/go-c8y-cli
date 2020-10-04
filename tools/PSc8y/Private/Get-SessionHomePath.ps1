Function Get-SessionHomePath {
    [cmdletbinding()]
    Param()

    $c8yFolderName = ".cumulocity"

    if ($env:C8Y_SESSION_HOME) {
        $HomePath = $env:C8Y_SESSION_HOME
    }
    elseif ($HOME) {
        # Use PS Automatic Variable
        $HomePath = Join-Path $HOME -ChildPath $c8yFolderName
    }
    else {
        # Check if on windows (PS 5.1)
        $IsWindowsOS = !($IsMacOS -or $IsLinux)
        if ($IsWindowsOS -and $env:HOMEDRIVE -and $env:HOMEPATH) {
            $HomePath = Join-Path -Path "$Env:HOMEDRIVE\$Env:HOMEPATH" -ChildPath $c8yFolderName
        } else {
            # default to current directory
            $HomePath = Join-Path "." -ChildPath $c8yFolderName
        }
    }

    $HomePath
}
