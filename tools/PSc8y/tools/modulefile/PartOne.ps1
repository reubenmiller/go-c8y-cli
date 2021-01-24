
#########################################################################################
#
# PSc8y Module
#
#########################################################################################

#Microsoft.PowerShell.Core\Set-StrictMode -Version Latest

#region script variables

$script:IsWindows = (-not (Get-Variable -Name IsWindows -ErrorAction Ignore)) -or $IsWindows
$script:IsLinux = (Get-Variable -Name IsLinux -ErrorAction Ignore) -and $IsLinux
$script:IsMacOS = (Get-Variable -Name IsMacOS -ErrorAction Ignore) -and $IsMacOS
$script:IsCoreCLR = $PSVersionTable.ContainsKey('PSEdition') -and $PSVersionTable.PSEdition -eq 'Core'
$script:Dependencies = Join-Path -Path $PSScriptRoot -ChildPath "Dependencies"

#endregion
