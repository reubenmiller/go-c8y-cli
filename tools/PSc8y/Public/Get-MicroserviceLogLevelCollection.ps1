# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-MicroserviceLogLevelCollection {
<#
.SYNOPSIS
List log levels of microservice

.DESCRIPTION
List all log levels of microservice.
(This only works for Spring Boot microservices based on Cumulocity Java Microservice SDK)


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/microservices_loglevels_list

.EXAMPLE
PS> Get-MicroserviceLogLevelCollection -Name my-microservice

List log levels of microservice


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Microservice name (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Name
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "microservices loglevels list"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y microservices loglevels list $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y microservices loglevels list $c8yargs
        }
    }

    End {}
}
