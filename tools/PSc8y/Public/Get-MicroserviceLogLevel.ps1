# Code generated from specification version 1.0.0: DO NOT EDIT
Function Get-MicroserviceLogLevel {
<#
.SYNOPSIS
Get log level of microservice

.DESCRIPTION
Get configured log level for a package (incl. subpackages), or a specific class.
(This only works for Spring Boot microservices based on Cumulocity Java Microservice SDK)


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/microservices_loglevels_get

.EXAMPLE
PS> Get-MicroserviceLogLevel -Name my-microservice -LoggerName org.example

Get log level of microservice for a package


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Microservice name (required)
        [Parameter(Mandatory = $true)]
        [object[]]
        $Name,

        # Name of the logger: Qualified name of package or class (required)
        [Parameter(Mandatory = $true)]
        [string]
        $LoggerName
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Get"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "microservices loglevels get"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y microservices loglevels get $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y microservices loglevels get $c8yargs
        }
    }

    End {}
}
