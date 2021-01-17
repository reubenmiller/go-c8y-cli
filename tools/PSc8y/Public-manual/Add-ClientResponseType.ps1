Function Add-ClientResponseType {
    <#
    .SYNOPSIS
    Add a PowerShell type to a Cumulocity object

    .DESCRIPTION
    This allows a custom type name to be given to powershell objects, so that the view formatting can be used (i.e. .ps1xml)

    .EXAMPLE
    $data | Add-ClientResponseType -Type "customType1"

    Add a type `customType1` to the input object

    .INPUTS
    Object[]

    .OUTPUTS
    Object[]
    #>
    [cmdletbinding()]
    Param(
        # Object to add the type name to
        [Parameter(
            Mandatory = $true,
            ValueFromPipeline = $true,
            ValueFromPipelineByPropertyName = $true)]
        [Object[]]
        $InputObject,

        # Type name to assign to the input objects
        [AllowNull()]
        [AllowEmptyString()]
        [Parameter(
            Mandatory = $true,
            Position = 1)]
        [string]
        $Type
    )

    Process {
        # check if raw values are being used (by statistics?), if not then add normal
        # look for value: switch on inventory, alarms, auditRecords, events, etc.
        if ($InputObject.statistics) {
            $TypeDefinitions = @(
                @{
                    property = "alarms"
                    type = "application/vnd.com.nsn.cumulocity.alarmCollection+json"
                },
                @{
                    property = "events"
                    type = "application/vnd.com.nsn.cumulocity.eventCollection+json"
                },
                @{
                    property = "managedObjects"
                    type = "application/vnd.com.nsn.cumulocity.managedObjectCollection+json"
                },
                @{
                    property = "operations"
                    type = "application/vnd.com.nsn.cumulocity.operationCollection+json"
                },
                @{
                    property = "auditRecords"
                    type = "application/vnd.com.nsn.cumulocity.auditRecordCollection+json"
                },
                @{
                    property = "applications"
                    type = "application/vnd.com.nsn.cumulocity.applicationCollection+json"
                },
                @{
                    property = "bulkOperations"
                    type = "application/vnd.com.nsn.cumulocity.bulkOperationCollection+json"
                },
                @{
                    property = "users"
                    type = "application/vnd.com.nsn.cumulocity.userCollection+json"
                }
            )

            $RawType = $null
            foreach ($iTypeDefinition in $TypeDefinitions) {
                if ($InputObject.$($iTypeDefinition.property)) {
                    $RawType = $iTypeDefinition.type
                    break
                }
            }

            if ($RawType) {
                $InputObject | Add-PowershellType -Type $RawType
                return
            }
        }

        foreach ($InObject in $InputObject) {
            if (-Not [string]::IsNullOrWhiteSpace($Type)) {
                [void]$InObject.PSObject.TypeNames.Insert(0, $Type)
            }
            $InObject
        }
    }
}
