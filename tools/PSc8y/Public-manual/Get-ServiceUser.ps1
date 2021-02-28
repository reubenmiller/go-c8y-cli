Function Get-ServiceUser {
    <#
    .SYNOPSIS
    Get service user

    .DESCRIPTION
    Get the service user associated to a microservice

    .LINK
    c8y microservices getServiceUser

    .EXAMPLE
    PS> Get-ServiceUser -Id $App.name

    Get application service user

    #>
    [cmdletbinding(SupportsShouldProcess = $true,
        PositionalBinding = $true,
        HelpUri = '',
        ConfirmImpact = 'None')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Microservice id (required)
        [Parameter(Mandatory = $true,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true)]
        [object[]]
        $Id
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Collection"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "microservices getServiceUser" -Exclude "Id"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type            = "application/vnd.com.nsn.cumulocity.applicationUserCollection+json"
            ItemType        = "application/vnd.com.nsn.cumulocity.bootstrapuser+json"
            BoundParameters = $PSBoundParameters
        }
    }

    Process {
        foreach ($item in (PSc8y\Expand-Microservice $Id)) {
            if ($item) {
                $appId = if ($item.id) { $item.id } else { $item }
            }

            if ($ClientOptions.ConvertToPS) {
                c8y microservices getServiceUser --id $appId $c8yargs `
                | ConvertFrom-ClientOutput @TypeOptions
            }
            else {
                c8y microservices getServiceUser --id $appId $c8yargs
            }
        }
    }

    End {}
}
