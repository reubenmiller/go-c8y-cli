Function Expand-PaginationObject {
<#
.SYNOPSIS
Expand a Cumulocity pagination result

.DESCRIPTION
Iterate through a Cumulocity pagination result set, and keep fetching the results
until the last page is found.

The cmdlet will only return once the total result set has been fetched, and the
items will be returned in one array.

.EXAMPLE
Invoke-ClientRequest -Uri "/inventory/managedObjects" -QueryParameters @{ pageSize = 2000 } -Raw | ConvertFrom-Json | Expand-PaginationObject

Get all managed objects in the platform (rest requests will be done in chunks of 2000)

.EXAMPLE
$data = Get-MeasurementCollection -Device testDevice -Raw -PageSize 2000 | Expand-PaginationObject

Get a measurement collection, then retrieve all the measurements by iterating through the pagination object

#>
  [cmdletbinding()]
  Param(
    # Response from a Cumulocity rest request. It must have the next property.
    [Parameter(
        Mandatory=$true,
        ValueFromPipeline=$true,
        ValueFromPipelineByPropertyName=$true)]
    [object]
    $InputObject,

    # Maximum number of pages to retrieve. If Zero or less, then it will retrieve all of the results
    [int]
    $MaxPage = 0
  )
  Begin {
    $InputCollection = New-Object System.Collections.ArrayList

    Function Get-ResultProperty {
      Param(
        [object] $InputObject
      )
      # Detect the type of c8y object
      $Prop = $null
      if ($null -ne $InputObject.managedObjects) {
        $Prop = "managedObjects"
      } elseif ($null -ne $InputObject.operations) {
        $Prop = "operations"
      } elseif ($null -ne $InputObject.alarms) {
        $Prop = "alarms"
      } elseif ($null -ne $InputObject.measurements) {
        $Prop = "measurements"
      } elseif ($null -ne $InputObject.events) {
        $Prop = "events"
      } elseif ($null -ne $InputObject.auditRecords) {
        $Prop = "auditRecords"
      } elseif ($null -ne $InputObject.connectors) {
        $Prop = "connectors"
      } elseif ($null -ne $InputObject.newDeviceRequests) {
        $Prop = "newDeviceRequests"
      } elseif ($null -ne $InputObject.externalIds) {
        $Prop = "externalIds"
      } elseif ($null -ne $InputObject.retentionRules) {
        $Prop = "retentionRules"
      } elseif ($null -ne $InputObject.groups) {
        $Prop = "groups"
      } elseif ($null -ne $InputObject.roles) {
        $Prop = "roles"
      } elseif ($null -ne $InputObject.users) {
        $Prop = "users"
      } elseif ($null -ne $InputObject.references) {
        $Prop = "references"
      } elseif ($null -ne $InputObject.tenants) {
        $Prop = "tenants"
      } elseif ($null -ne $InputObject.options) {
        $Prop = "options"
      }

      $Prop
    }
  }

  Process {
    $null = $InputCollection.Add($InputObject)
  }

  End {
    $ProcessObject = $InputCollection | Select-Object -First 1

    Write-Verbose "Input Collection Count: $($InputCollection.Count)"

    if (($InputCollection.Count -eq 0) -or (($InputCollection.Count -gt 1) -and !$InputCollection[0].next)) {
      Write-Warning "Input object is not a Cumulocity Pagination Object"
      $ProcessObject
      return;
    }

    $ResultCollection = New-Object System.Collections.ArrayList

    $Result = $ProcessObject

    $Prop = Get-ResultProperty -InputObject $Result
    if ($Result.$Prop -is [array]) {
      $null = $ResultCollection.AddRange($Result.$Prop)
    } else {
      $null = $ResultCollection.Add($Result.$Prop)
    }

    $Iteration = 1

    $Done = $false

    if ($Result.next) {
      while (!$Done) {
        Write-Verbose "Requesting next item: (iteration $Iteration)"

        $Result = Invoke-ClientRequest -Method Get -Uri $Result.next -Raw -AsPSObject:$false | ConvertFrom-Json

        # Detect the type of c8y object
        $Prop = Get-ResultProperty -InputObject $Result

        if ($null -eq $Prop)
        {
          Write-Warning "Could not find the array object property. Only alarms, events, managedObjects, measurments and operations properties are supported"
        }
        else
        {
          Write-Verbose "Found array object property [$Prop]"
        }

        if ($Result -and $Prop) {
          if ($Result.$Prop -is [array]) {
            $null = $ResultCollection.AddRange($Result.$Prop)
          } else {
            $null = $ResultCollection.Add($Result.$Prop)
          }
        }

        $Done = !$Result -or !$Result.next -or ($Result.$Prop.Count -eq 0) `
          -or ($MaxPage -gt 0 -and $Iteration -ge $MaxPage) `
          -or ($Result.$Prop.Count -lt $Result.statistics.pageSize)
        $Iteration += 1
      }
    }

    $ResultCollection
  }
}
