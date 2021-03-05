# Code generated from specification version 1.0.0: DO NOT EDIT
Function Remove-ApplicationBinary {
<#
.SYNOPSIS
Delete application binary

.DESCRIPTION
Remove an application binaries related to the given application
The active version can not be deleted and the server will throw an error if you try.


.LINK
c8y applications deleteApplicationBinary

.EXAMPLE
PS> Remove-ApplicationBinary -Application $app.id -BinaryId $appBinary.id

Remove an application binary related to a Hosted (web) application

.EXAMPLE
PS> Get-ApplicationBinaryCollection -Id $app.id | Remove-ApplicationBinary -Application $app.id

Remove all application binaries (except for the active one) for an application (using pipeline)


#>
    [cmdletbinding(SupportsShouldProcess = $true,
                   PositionalBinding=$true,
                   HelpUri='',
                   ConfirmImpact = 'High')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Application id (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Application,

        # Application binary id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [Alias("id")]
        [string[]]
        $BinaryId
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Delete"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "applications deleteApplicationBinary"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = ""
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        $Force = if ($PSBoundParameters.ContainsKey("Force")) { $PSBoundParameters["Force"] } else { $False }
        if (!$Force -and !$WhatIfPreference) {
            $items = (PSc8y\Expand-Id $BinaryId)

            $shouldContinue = $PSCmdlet.ShouldProcess(
                (PSc8y\Get-C8ySessionProperty -Name "tenant"),
                (Format-ConfirmationMessage -Name $PSCmdlet.MyInvocation.InvocationName -InputObject $items)
            )
            if (!$shouldContinue) {
                return
            }
        }

        if ($ClientOptions.ConvertToPS) {
            $BinaryId `
            | Group-ClientRequests `
            | c8y applications deleteApplicationBinary $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $BinaryId `
            | Group-ClientRequests `
            | c8y applications deleteApplicationBinary $c8yargs
        }
        
    }

    End {}
}
