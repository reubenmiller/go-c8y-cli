# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-MicroserviceBinary {
<#
.SYNOPSIS
Create microservice binary

.DESCRIPTION
Create/upload a new microservice binary. For the applications of type 'MICROSERVICE' to be available for Cumulocity platform users, a binary zip file must be uploaded.
For the microservice application, the zip file must consist of    * cumulocity.json - file describing the deployment
    * image.tar - executable docker image

For the web application, the zip file must include index.html in the root directory.


.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/microservices_createBinary

.EXAMPLE
PS> New-MicroserviceBinary -Id $App.id -File $MicroserviceZip

Upload microservice binary


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Microservice id (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Id,

        # File to be uploaded as a binary (required)
        [Parameter(Mandatory = $true)]
        [string]
        $File
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "microservices createBinary"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.managedObject+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Id `
            | Group-ClientRequests `
            | c8y microservices createBinary $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Id `
            | Group-ClientRequests `
            | c8y microservices createBinary $c8yargs
        }
        
    }

    End {}
}
