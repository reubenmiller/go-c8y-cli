Function Add-PowershellType {
<#
.SYNOPSIS
Add a powershell type name to a powershell object

.DESCRIPTION
This allows a custom type name to be given to powershell objects, so that the view formatting can be used (i.e. .ps1xml)

.EXAMPLE
$data | Add-PowershellType -Type "customType1"

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
    foreach ($InObject in $InputObject) {
      if (-Not [string]::IsNullOrWhiteSpace($Type)) {
        [void]$InObject.PSObject.TypeNames.Insert(0, $Type)
      }
      $InObject
    }
  }
}
