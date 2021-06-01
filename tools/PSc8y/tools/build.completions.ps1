[cmdletbinding()]
param(
    # Module name
    [Parameter()]
    [string]
    $ModuleName = "PSc8y",

    # Output file
    [Parameter()]
    [string]
    $OutputFile = "$PSScriptRoot/../utilities/register-completions.ps1"
)

[array]$Completions = @(
    # Tenant Option
    @{ name = "*TenantOption*"; ParameterName = "Category"; completion = '$script:CompleteTenantOptionCategory' }
    @{ name = "*TenantOption*"; ParameterName = "Key"; completion = '$script:CompleteTenantOptionKey' }
        
    # User Group
    @{ name = "*UserGroup*"; ParameterName = "Id"; completion = '$script:CompleteUserGroup' }

    # Application
    @{ name = "*Application"; ParameterName = "Id"; completion = '$script:CompleteApplication' }

    # Microservice
    @{ name = "*Microservice"; ParameterName = "Id"; completion = '$script:CompleteMicroservice' }

    # Tenant
    @{ name = "*Tenant"; ParameterName = "Id"; completion = '$script:CompleteTenant' }

    # User
    @{ name = "*-User"; ParameterName = "Id"; completion = '$script:CompleteUser' }
    @{ name = "*"; ParameterName = "User"; completion = '$script:CompleteUser' }

    # DeviceGroup
    @{ name = "*-DeviceGroup*"; ParameterName = "Id"; completion = '$script:CompleteDeviceGroup' }

    # Measurement
    @{ name = "*-Measurement*"; ParameterName = "ValueFragmentType"; completion = '$script:CompleteMeasurementFragmentType' }
    @{ name = "*-Measurement*"; ParameterName = "ValueFragmentSeries"; completion = '$script:CompleteMeasurementSeries' }
    @{ name = "*-Measurement*"; ParameterName = "Series"; completion = '$script:CompleteMeasurementFullSeries' }

    # UserGroup
    @{ name = "Add-UserToGroup"; ParameterName = "Group"; completion = '$script:CompleteUserGroup' }
    @{ name = "Add-RoleToGroup"; ParameterName = "Group"; completion = '$script:CompleteUserGroup' }

    # Add
    @{ name = "*AssetToGroup"; ParameterName = "*Group*"; completion = '$script:CompleteDeviceGroup' }

    # Device
    @{ name = "*"; ParameterName = "*Device*"; completion = '$CompleteDevice' }

    # Agent
    @{ name = "*"; ParameterName = "*Agent*"; completion = '$CompleteAgent' }

    # Tenant
    @{ name = "*"; ParameterName = "Tenant"; completion = '$script:CompleteTenant' }

    # Role
    @{ name = "*"; ParameterName = "Role"; completion = '$script:CompleteRole' }
    @{ name = "*"; ParameterName = "Roles"; completion = '$script:CompleteRole' }
)

# build a hashtable of command names and parameters to make it quicker to lookup
[array] $AllParameters = Get-Command -Module $ModuleName | ForEach-Object {
    $command = $_
    foreach ($key in @($command.Parameters.Keys)) {
        $id = $command.Name + ":" + $key
        Write-Verbose "id: $id"
        [pscustomobject]@{ id = $id; command = $command.Name; parameter = $key }
    }
}

if ($AllParameters.Count -gt 0) {
    Write-Verbose "Total Parameters: $($AllParameters.Count)"
}

$builder = New-Object System.Text.StringBuilder

foreach ($item in $Completions) {
    $ParamName = $item.ParameterName
    $CompletionName = $item.completion
    $query = $item.name + ":" + $ParamName
    [array] $CommandNames = $AllParameters | Where-Object id -like $query | Select-Object -ExpandProperty command -Unique

    Write-Verbose "Found matching query ($query): $($CommandNames.Count)"

    if ($CommandNames.Count -gt 0) {
        $VarName = "$($item.name)$ParamName" -replace "[^a-z0-9]", "" 
        [void]$builder.AppendLine(@"
`$CommandsWith$VarName = @(
    $("`"" + ($CommandNames -join "`",`n    `"") + "`"")
)
Register-ArgumentCompleter -CommandName `$CommandsWith$VarName -ParameterName $ParamName -ScriptBlock $CompletionName

"@)
    }
}

$builder.ToString() | Out-File -FilePath $OutputFile
