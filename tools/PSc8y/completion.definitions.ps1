<# 
.SYNOPSIS
Provides argument completions for dynamic data by querying Cumulocity for the values
#>
$RoleCompleter = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-RoleCollection -PageSize 100 `
    | Select-Object -ExpandProperty id `
    | Where-Object { $_ -like "$searchFor*" } `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_,
            $_,
            'ParameterValue',
            "$_"
        )
    }
}

$TenantCompleter = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-TenantCollection -PageSize 100 -WarningAction SilentlyContinue `
    | Where-Object { $_.id -like "$searchFor*" } `
    | ForEach-Object {
        $id = $_.id
        $details = ("{0} ({1})" -f $_.id, ($_.domain -split "\.")[0])
        [System.Management.Automation.CompletionResult]::new(
            $id,
            $details,
            'ParameterValue',
            $id
        )
    }
}

$ApplicationCompleter = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-ApplicationCollection -PageSize 100 -WarningAction SilentlyContinue `
    | Where-Object { $_.id -like "$searchFor*" -or $_.name -like "$searchFor*" } `
    | ForEach-Object {
        $value = $_.name
        $details = ("{0} ({1})" -f $_.id, $_.name)
        [System.Management.Automation.CompletionResult]::new(
            $value,
            $details,
            'ParameterValue',
            $value
        )
    }
}

$MicroserviceCompleter = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-ApplicationCollection -PageSize 100 -WarningAction SilentlyContinue `
    | Where-Object { $_.type -eq "MICROSERVICE" } `
    | Where-Object { $_.id -like "$searchFor*" -or $_.name -like "$searchFor*" } `
    | ForEach-Object {
        $details = ("{0} ({1})" -f $_.id, $_.name)
        [System.Management.Automation.CompletionResult]::new(
            $_.id,
            $details,
            'ParameterValue',
            $_.id
        )
    }
}

$UserGroupCompleter = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-UserGroupCollection -PageSize 100 -WarningAction SilentlyContinue `
    | Where-Object { $_.id -like "$searchFor*" -or $_.name -like "$searchFor*" } `
    | ForEach-Object {
        $value = $_.name
        $details = ("{0} ({1})" -f $_.id, $_.name)
        [System.Management.Automation.CompletionResult]::new(
            $_.id,
            $details,
            'ParameterValue',
            $_.id
        )
    }
}

$TenantOptionCategoryCompleter = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-TenantOptionCollection -PageSize 100 -WarningAction SilentlyContinue `
    | Where-Object { $_.category -like "$searchFor*" } `
    | Select-Object -Unique -ExpandProperty category `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_,
            $_,
            'ParameterValue',
            $_
        )
    }
}

$TenantOptionKeyCompleter = {
    param ($commandName, $parameterName, $wordToComplete, $ast, $fakeBoundParameters)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    $options = Get-TenantOptionForCategory -Category $fakeBoundParameters["category"] -PageSize 100 -WarningAction SilentlyContinue
    
    if ($options -is [hashtable]) {
        $keys = $options.keys
    } else {
        $keys = $options.psobject.Properties.Name
    }
    
    $keys `
    | Where-Object { $_ -like "$searchFor*" } `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_,
            $_,
            'ParameterValue',
            $_
        )
    }
}

$UserCompleter = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-UserCollection -PageSize 1000 -WarningAction SilentlyContinue `
    | Where-Object { $_.id -like "$searchFor*" -or $_.email -like "$searchFor*" } `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_.id,
            $_.id,
            'ParameterValue',
            $_.id
        )
    }
}

$DeviceGroupCompleter = {
    param ($commandName, $parameterName, $wordToComplete)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    Get-DeviceGroupCollection -Name "$searchFor*" -PageSize 100 -WarningAction SilentlyContinue `
    | Where-Object { $_.name -like "$searchFor*" } `
    | Sort-Object { $_.name } `
    | ForEach-Object {
        $details = ("{0} ({1})" -f $_.name, $_.id)
        [System.Management.Automation.CompletionResult]::new(
            $_.id,
            $details,
            'ParameterValue',
            $details
        )
    }
}

$MeasurementFragmentTypeCompleter = {
    param ($commandName, $parameterName, $wordToComplete, $ast, $fakeBoundParameters)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    if (!$fakeBoundParameters.ContainsKey("Device")) {
        return
    }

    $device = $fakeBoundParameters["Device"]

    Get-SupportedMeasurements -Device:$device -WarningAction SilentlyContinue `
    | Where-Object { $_ -like "$searchFor*" } `
    | Sort-Object { $_ } `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_,
            $_,
            'ParameterValue',
            $_
        )
    }
}

$MeasurementSeriesCompleter = {
    param ($commandName, $parameterName, $wordToComplete, $ast, $fakeBoundParameters)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    if (!$fakeBoundParameters.ContainsKey("Device")) {
        return
    }

    $device = $fakeBoundParameters["Device"]
    $valueFragmentType = $fakeBoundParameters["valueFragmentType"]
    
    Get-SupportedSeries -Device:$device -WarningAction SilentlyContinue `
    | Where-Object { $_ -like "$valueFragmentType.*" } `
    | Where-Object { $_ -like "*.$searchFor*" } `
    | ForEach-Object { "$_".Split(".")[-1] } `
    | Sort-Object { $_ } `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_,
            $_,
            'ParameterValue',
            $_
        )
    }
}

$MeasurementFullSeriesCompleter = {
    param ($commandName, $parameterName, $wordToComplete, $ast, $fakeBoundParameters)
    if ($wordToComplete -is [array]) {
        $searchFor = $wordToComplete | Select-Object -Last 1
    } else {
        $searchFor = $wordToComplete
    }

    if (!$fakeBoundParameters.ContainsKey("Device")) {
        return
    }

    $device = $fakeBoundParameters["Device"]
    
    Get-SupportedSeries -Device:$device -WarningAction SilentlyContinue `
    | Where-Object { $_ -like "$searchFor*" } `
    | Sort-Object { $_ } `
    | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new(
            $_,
            $_,
            'ParameterValue',
            $_
        )
    }
}

# Tenant options
Register-ArgumentCompleter -CommandName Update-TenantOption -ParameterName category -ScriptBlock $TenantOptionCategoryCompleter
Register-ArgumentCompleter -CommandName Update-TenantOption -ParameterName key -ScriptBlock $TenantOptionKeyCompleter
Register-ArgumentCompleter -CommandName Remove-TenantOption -ParameterName category -ScriptBlock $TenantOptionCategoryCompleter
Register-ArgumentCompleter -CommandName Remove-TenantOption -ParameterName key -ScriptBlock $TenantOptionKeyCompleter
Register-ArgumentCompleter -CommandName Update-TenantOptionEditable -ParameterName category -ScriptBlock $TenantOptionCategoryCompleter
Register-ArgumentCompleter -CommandName Update-TenantOptionEditable -ParameterName key -ScriptBlock $TenantOptionKeyCompleter

# User Group
Register-ArgumentCompleter -CommandName Get-UserGroup -ParameterName Id -ScriptBlock $UserGroupCompleter
Register-ArgumentCompleter -CommandName Update-UserGroup -ParameterName Id -ScriptBlock $UserGroupCompleter
Register-ArgumentCompleter -CommandName Remove-UserGroup -ParameterName Id -ScriptBlock $UserGroupCompleter

# Service User (microservice)
Register-ArgumentCompleter -CommandName New-ServiceUser -ParameterName Roles -ScriptBlock $RoleCompleter

# Application
Register-ArgumentCompleter -CommandName Get-Application -ParameterName Id -ScriptBlock $ApplicationCompleter
Register-ArgumentCompleter -CommandName Update-Application -ParameterName Id -ScriptBlock $ApplicationCompleter
Register-ArgumentCompleter -CommandName Remove-Application -ParameterName Id -ScriptBlock $ApplicationCompleter

# Microservice
Register-ArgumentCompleter -CommandName Get-Microservice -ParameterName Id -ScriptBlock $MicroserviceCompleter
Register-ArgumentCompleter -CommandName Update-Microservice -ParameterName Id -ScriptBlock $MicroserviceCompleter
Register-ArgumentCompleter -CommandName Remove-Microservice -ParameterName Id -ScriptBlock $MicroserviceCompleter

# tenant
Register-ArgumentCompleter -CommandName Get-Tenant -ParameterName Id -ScriptBlock $TenantCompleter
Register-ArgumentCompleter -CommandName Update-Tenant -ParameterName Id -ScriptBlock $TenantCompleter
Register-ArgumentCompleter -CommandName Remove-Tenant -ParameterName Id -ScriptBlock $TenantCompleter

# user
Register-ArgumentCompleter -CommandName Get-User -ParameterName Id -ScriptBlock $UserCompleter
Register-ArgumentCompleter -CommandName Update-User -ParameterName Id -ScriptBlock $UserCompleter
Register-ArgumentCompleter -CommandName Remove-User -ParameterName Id -ScriptBlock $UserCompleter

# device group
Register-ArgumentCompleter -CommandName Get-DeviceGroup -ParameterName Id -ScriptBlock $DeviceGroupCompleter
Register-ArgumentCompleter -CommandName Update-DeviceGroup -ParameterName Id -ScriptBlock $DeviceGroupCompleter
Register-ArgumentCompleter -CommandName Remove-DeviceGroup -ParameterName Id -ScriptBlock $DeviceGroupCompleter

# measurement
Register-ArgumentCompleter -CommandName Get-MeasurementCollection -ParameterName valueFragmentType -ScriptBlock $MeasurementFragmentTypeCompleter
Register-ArgumentCompleter -CommandName Get-MeasurementCollection -ParameterName valueFragmentSeries -ScriptBlock $MeasurementSeriesCompleter
Register-ArgumentCompleter -CommandName Get-MeasurementSeries -ParameterName Series -ScriptBlock $MeasurementFullSeriesCompleter


# User group
Register-ArgumentCompleter -CommandName Add-UserToGroup -ParameterName User -ScriptBlock $UserCompleter
Register-ArgumentCompleter -CommandName Add-UserToGroup -ParameterName Group -ScriptBlock $UserGroupCompleter

# Add Role to Group
Register-ArgumentCompleter -CommandName Add-RoleToGroup -ParameterName Group -ScriptBlock $UserGroupCompleter
Register-ArgumentCompleter -CommandName Add-RoleToGroup -ParameterName Role -ScriptBlock $RoleCompleter

# Add Role to User
Register-ArgumentCompleter -CommandName Add-RoleToUser -ParameterName User -ScriptBlock $UserCompleter
Register-ArgumentCompleter -CommandName Add-RoleToUser -ParameterName Role -ScriptBlock $RoleCompleter

Register-ArgumentCompleter -CommandName Get-UserGroupMembershipCollection -ParameterName Id -ScriptBlock $UserGroupCompleter

$ModuleName = "PSc8y"

Function Register-C8yArgumentCompletion {
    [cmdletbinding()]
    Param(
        [string] $Module,
        [string] $ArgumentName,
        [scriptblock] $Completer
    )

    if (!(Get-Command -Name Register-ArgumentCompleter -ErrorAction SilentlyContinue)) {
        Write-Warning "Register-ArgumentCompleter is required to use this function"
        return
    }

    $commands = Get-Command -Module $Module -CommandType "Function" `
        | Where-Object {
            $null -ne $_.Parameters -and $_.Parameters.ContainsKey($ArgumentName)
        }
    Register-ArgumentCompleter -CommandName $commands -ParameterName $ArgumentName -ScriptBlock $Completer
}

Function Get-c8yCommand {
    [cmdletbinding()]
    Param(
        [string] $Module,
        [string] $IncludeCommand,
        [string] $ExcludeCommand,
        [string] $ParameterName
    )

    Get-Command -Module $Module -CommandType "Function" `
        | Where-Object {          
            ([string]::IsNullOrWhiteSpace($IncludeCommand) -or $_.Name -like $IncludeCommand) `
                -and ([string]::IsNullOrWhiteSpace($ExcludeCommand) -or $_.Name -notlike $ExcludeCommand) `
                -and ($null -ne $_.Parameters) `
                -and ($_.Parameters.ContainsKey($ParameterName))
        }
}

# Tenant/s parameters
Register-C8yArgumentCompletion -Module $ModuleName -ArgumentName "Tenant" -Completer $TenantCompleter
Register-C8yArgumentCompletion -Module $ModuleName -ArgumentName "Tenants" -Completer $TenantCompleter

Register-C8yArgumentCompletion -Module $ModuleName -ArgumentName "User" -Completer $TenantCompleter
Register-C8yArgumentCompletion -Module $ModuleName -ArgumentName "Role" -Completer $RoleCompleter
Register-C8yArgumentCompletion -Module $ModuleName -ArgumentName "Roles" -Completer $RoleCompleter
