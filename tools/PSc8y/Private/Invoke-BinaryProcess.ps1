##############################################################################
##
## Invoke-BinaryProcess
##
## From Windows PowerShell Cookbook (O'Reilly)
## by Lee Holmes (http://www.leeholmes.com/guide)
# https://www.powershellcookbook.com/recipe/WCiL/capture-and-redirect-binary-process-output
##
##############################################################################
Function Invoke-BinaryProcess {
<#

.SYNOPSIS

Invokes a process that emits or consumes binary data.

.EXAMPLE

PS > Invoke-BinaryProcess binaryProcess.exe -RedirectOutput -ArgumentList "-Emit" |
       Invoke-BinaryProcess binaryProcess.exe -RedirectInput -ArgumentList "-Consume"

#>
    [cmdletbinding(
        SupportsShouldProcess = $true
    )]
    param(
        ## The name of the process to invoke
        [string] $ProcessName,

        ## Specifies that input to the process should be treated as
        ## binary
        [Alias("Input")]
        [switch] $RedirectInput,

        ## Specifies that the output of the process should be treated
        ## as binary
        [Alias("Output")]
        [switch] $RedirectOutput,

        # Redirect stderr
        [switch] $RedirectStdError,

        [switch]
        $AsText,

        ## Specifies the arguments for the process
        [string] $ArgumentList
    )

    Set-StrictMode -Version 3

    ## Prepare to invoke the process
    $processStartInfo = New-Object System.Diagnostics.ProcessStartInfo
    $processStartInfo.FileName = (Get-Command $processname).Definition
    $processStartInfo.WorkingDirectory = (Get-Location).Path
    if ($argumentList) { $processStartInfo.Arguments = $argumentList }
    $processStartInfo.UseShellExecute = $false

    ## Always redirect the input and output of the process.
    ## Sometimes we will capture it as binary, other times we will
    ## just treat it as strings.
    $processStartInfo.RedirectStandardOutput = $true
    $processStartInfo.RedirectStandardInput = $true

    if ($RedirectStdError) {
        $processStartInfo.RedirectStandardError = $true
    }

    $process = [System.Diagnostics.Process]::Start($processStartInfo)

    ## If we've been asked to redirect the input, treat it as bytes.
    ## Otherwise, write any input to the process as strings.
    if ($redirectInput) {
        $inputBytes = @($input)
        $process.StandardInput.BaseStream.Write($inputBytes, 0, $inputBytes.Count)
        $process.StandardInput.Close()
    }
    else {
        $input | ForEach-Object { $process.StandardInput.WriteLine($_) }
        $process.StandardInput.Close()
    }

    ## If we've been asked to redirect the output, treat it as bytes.
    ## Otherwise, read any input from the process as strings.
    if ($redirectOutput) {

        while (!$process.StandardOutput.EndOfStream)
        {
            $line = $process.StandardOutput.ReadLine();

            if ($AsText) {
                $line
            } else {
                if ($null -ne $line) {
                    ConvertFrom-Json -Depth 100 -InputObject $line
                }
            }
        }

        # $byteRead = -1
        # do {
        #     $byteRead = $process.StandardOutput.BaseStream.ReadByte()
        #     if ($byteRead -ge 0) { $byteRead }
        # } while ($byteRead -ge 0)
    }
    else {
        $process.StandardOutput.ReadToEnd()
    }

    if ($RedirectStdError) {
        $year = Get-Date -Format "yyyy"
        while (!$process.StandardError.EndOfStream) {
            $line = $process.StandardError.ReadLine()

            if ($line.Contains("What If")) {
                # remove the timestamp (if present)
                $line = $line -replace ".*(What if:)", '$1'
                Write-Host $line -ForegroundColor "Green"
            }
            elseif ($line.StartsWith("Error")) {
                Write-Verbose $line
            }
            elseif (!$line.StartsWith($year)) {
                Write-Host $line -ForegroundColor "Green"
            }
            else {
                # Normal verbose message
                Write-Verbose $line
            }
        }
    }

    # Must wait for exit before returning!
    $process.WaitForExit()

    # Set exit code so it is propagated
    Write-Verbose ("Exit code: {0}" -f $process.ExitCode)
    $global:LASTEXITCODE = $process.ExitCode
    $global:C8Y_EXITCODE = $process.ExitCode
    if ($process.ExitCode -ne 0) {
        Write-Error ("c8y returned a non-zero exit code. code={0}" -f $process.ExitCode)
    }
}