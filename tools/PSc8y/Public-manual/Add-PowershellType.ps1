Function Add-PowershellType {
<#
.SYNOPSIS
Add a powershell type name to a powershell object

.DESCRIPTION
This allows a custom type name to be given to powershell objects, so that the view formatting can be used (i.e. .ps1xml)

.EXAMPLE
$data | Add-PowershellType -Type "customType1"

.OUTPUTS Object[]
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
    [Parameter(
      Mandatory = $true,
      Position = 1)]
    [string]
    $Type
  )

  Process {
    foreach ($InObject in $InputObject) {
      [void]$InObject.PSObject.TypeNames.Insert(0, $Type)
      $InObject
    }
  }
}
