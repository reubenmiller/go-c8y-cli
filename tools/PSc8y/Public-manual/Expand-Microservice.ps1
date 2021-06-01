Function Expand-Microservice {
<#
.SYNOPSIS
Expand a list of microservices objects

.DESCRIPTION
Expand a list of microservices replacing any ids or names with the actual microservice object.

.NOTES
If the given object is already an microservice object, then it is added with no additional lookup

.PARAMETER InputObject
List of ids, names or microservice objects

.EXAMPLE
Expand-Microservice "app-name"

Retrieve the microservice objects by name or id

.EXAMPLE
Get-C8yMicroserviceCollection *app* | Expand-Microservice

Get all the microservice object (with app in their name). Note the Expand cmdlet won't do much here except for returning the input objects.

#>
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory=$true,
            ValueFromPipeline=$true,
            Position=0
        )]
        [object[]] $InputObject
    )

    Begin {
        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }
    }

    Process {
        [array] $AllMicroservices = foreach ($iApp in $InputObject)
        {
            # Already an app object, so do nothing
            if ($iApp.id) {
                $iApp
                continue
            }

            if ($iApp.applicationId) {
                PSc8y\Get-Microservice -Id $iApp.applicationId -Dry:$false -AsPSObject
                continue
            }

            if ("$iApp" -match "^\d+$") {
                # Provided with an id
                $iApp
            } else {
                # Provided with a query
                PSc8y\Get-MicroserviceCollection -PageSize 2000 -AsPSObject |
                        Where-Object { $_.name -like "$iApp" }
            }
        }

        $AllMicroservices
    }
}
