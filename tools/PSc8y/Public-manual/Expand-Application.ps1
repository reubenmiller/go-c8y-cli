Function Expand-Application {
<#
.SYNOPSIS
Expand a list of applications replacing any ids or names with the actual application object.

.DESCRIPTION
The list of applications will be expanded to include the full application representation by fetching
the data from Cumulocity.

.NOTES
If the given object is already an application object, then it is added with no additional lookup

.PARAMETER InputObject
List of ids, names or application objects

.PARAMETER Type
Limit the types of object by a specific type

.EXAMPLE
Expand-Application "app-name"

Retrieve the application objects by name or id

.EXAMPLE
Get-C8yApplicationCollection *app* | Expand-Application

Get all the application object (with app in their name). Note the Expand cmdlet won't do much here except for returning the input objects.

.EXAMPLE
Expand-Application * -Type MICROSERVICE

Expand applications that match a name of "*" and have a type of "MICROSERVICE"

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
        [array] $AllApplications = foreach ($iApp in $InputObject)
        {
            # Already an app object, so do nothing
            if ($iApp.id) {
                $iApp
                continue
            }

            if ($iApp.applicationId) {
                PSc8y\Get-Application -Id $iApp.applicationId -WhatIf:$false
                continue
            }

            if ("$iApp" -match "^\d+$") {
                # Provided with an id
                $iApp
                # PSc8y\Get-Application -Id $iApp -WhatIf:$false
            } else {
                # Provided with a query
                PSc8y\Get-ApplicationCollection -PageSize 2000 |
                        Where-Object { $_.name -like "$iApp" }
            }
        }

        $AllApplications
    }
}
