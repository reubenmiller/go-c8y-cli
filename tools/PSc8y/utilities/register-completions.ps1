$CommandsWithTenantOptionCategory = @(
    "Get-TenantOption",
    "Get-TenantOptionForCategory",
    "New-TenantOption",
    "Remove-TenantOption",
    "Update-TenantOption",
    "Update-TenantOptionBulk",
    "Update-TenantOptionEditable"
)
Register-ArgumentCompleter -CommandName $CommandsWithTenantOptionCategory -ParameterName Category -ScriptBlock $script:CompleteTenantOptionCategory

$CommandsWithTenantOptionKey = @(
    "Get-TenantOption",
    "New-TenantOption",
    "Remove-TenantOption",
    "Update-TenantOption",
    "Update-TenantOptionEditable"
)
Register-ArgumentCompleter -CommandName $CommandsWithTenantOptionKey -ParameterName Key -ScriptBlock $script:CompleteTenantOptionKey

$CommandsWithUserGroupId = @(
    "Get-UserGroup",
    "Get-UserGroupMembershipCollection",
    "Remove-UserGroup",
    "Update-UserGroup"
)
Register-ArgumentCompleter -CommandName $CommandsWithUserGroupId -ParameterName Id -ScriptBlock $script:CompleteUserGroup

$CommandsWithApplicationId = @(
    "Copy-Application",
    "Get-Application",
    "Remove-Application",
    "Update-Application"
)
Register-ArgumentCompleter -CommandName $CommandsWithApplicationId -ParameterName Id -ScriptBlock $script:CompleteApplication

$CommandsWithMicroserviceId = @(
    "Disable-Microservice",
    "Enable-Microservice",
    "Get-Microservice",
    "Remove-Microservice",
    "Update-Microservice"
)
Register-ArgumentCompleter -CommandName $CommandsWithMicroserviceId -ParameterName Id -ScriptBlock $script:CompleteMicroservice

$CommandsWithTenantId = @(
    "Get-Tenant",
    "Remove-Tenant",
    "Update-Tenant"
)
Register-ArgumentCompleter -CommandName $CommandsWithTenantId -ParameterName Id -ScriptBlock $script:CompleteTenant

$CommandsWithUserId = @(
    "Get-User",
    "Remove-User",
    "Update-User"
)
Register-ArgumentCompleter -CommandName $CommandsWithUserId -ParameterName Id -ScriptBlock $script:CompleteUser

$CommandsWithUser = @(
    "Add-RoleToUser",
    "Add-UserToGroup",
    "Get-AuditRecordCollection",
    "Get-RoleReferenceCollectionFromUser",
    "New-AuditRecord",
    "Remove-RoleFromUser",
    "Remove-UserFromGroup"
)
Register-ArgumentCompleter -CommandName $CommandsWithUser -ParameterName User -ScriptBlock $script:CompleteUser

$CommandsWithDeviceGroupId = @(
    "Get-DeviceGroup",
    "Remove-DeviceGroup",
    "Update-DeviceGroup"
)
Register-ArgumentCompleter -CommandName $CommandsWithDeviceGroupId -ParameterName Id -ScriptBlock $script:CompleteDeviceGroup

$CommandsWithMeasurementValueFragmentType = @(
    "Get-MeasurementCollection"
)
Register-ArgumentCompleter -CommandName $CommandsWithMeasurementValueFragmentType -ParameterName ValueFragmentType -ScriptBlock $script:CompleteMeasurementFragmentType

$CommandsWithMeasurementValueFragmentSeries = @(
    "Get-MeasurementCollection"
)
Register-ArgumentCompleter -CommandName $CommandsWithMeasurementValueFragmentSeries -ParameterName ValueFragmentSeries -ScriptBlock $script:CompleteMeasurementSeries

$CommandsWithMeasurementSeries = @(
    "Get-MeasurementSeries"
)
Register-ArgumentCompleter -CommandName $CommandsWithMeasurementSeries -ParameterName Series -ScriptBlock $script:CompleteMeasurementFullSeries

$CommandsWithAddUserToGroupGroup = @(
    "Add-UserToGroup"
)
Register-ArgumentCompleter -CommandName $CommandsWithAddUserToGroupGroup -ParameterName Group -ScriptBlock $script:CompleteUserGroup

$CommandsWithAddRoleToGroupGroup = @(
    "Add-RoleToGroup"
)
Register-ArgumentCompleter -CommandName $CommandsWithAddRoleToGroupGroup -ParameterName Group -ScriptBlock $script:CompleteUserGroup

$CommandsWithAssetToGroupGroup = @(
    "Add-AssetToGroup"
)
Register-ArgumentCompleter -CommandName $CommandsWithAssetToGroupGroup -ParameterName *Group* -ScriptBlock $script:CompleteDeviceGroup

$CommandsWithDevice = @(
    "Add-AssetToGroup",
    "Add-ChildDeviceToDevice",
    "Add-DeviceToGroup",
    "Get-AlarmCollection",
    "Get-ChildAssetCollection",
    "Get-ChildDeviceCollection",
    "Get-ChildDeviceReference",
    "Get-DeviceParent",
    "Get-EventCollection",
    "Get-ExternalIdCollection",
    "Get-MeasurementCollection",
    "Get-MeasurementSeries",
    "Get-OperationCollection",
    "Get-SupportedMeasurements",
    "Get-SupportedSeries",
    "Get-UserCollection",
    "New-Alarm",
    "New-Event",
    "New-ExternalId",
    "New-Measurement",
    "New-Operation",
    "New-TestAlarm",
    "New-TestDeviceGroup",
    "New-TestEvent",
    "New-TestMeasurement",
    "New-TestOperation",
    "Open-Website",
    "Remove-AlarmCollection",
    "Remove-AssetFromGroup",
    "Remove-ChildDeviceFromDevice",
    "Remove-DeviceFromGroup",
    "Remove-EventCollection",
    "Remove-MeasurementCollection",
    "Remove-OperationCollection",
    "Set-DeviceRequiredAvailability",
    "Update-AlarmCollection",
    "Watch-Alarm",
    "Watch-Event",
    "Watch-ManagedObject",
    "Watch-Measurement",
    "Watch-NotificationChannel",
    "Watch-Operation"
)
Register-ArgumentCompleter -CommandName $CommandsWithDevice -ParameterName *Device* -ScriptBlock $CompleteDevice

$CommandsWithAgent = @(
    "Get-DeviceCollection",
    "Get-OperationCollection",
    "New-TestDevice",
    "Remove-OperationCollection"
)
Register-ArgumentCompleter -CommandName $CommandsWithAgent -ParameterName *Agent* -ScriptBlock $CompleteAgent

$CommandsWithTenant = @(
    "Add-RoleToGroup",
    "Add-RoleToUser",
    "Add-UserToGroup",
    "Disable-Application",
    "Disable-Microservice",
    "Enable-Application",
    "Enable-Microservice",
    "Get-ApplicationReferenceCollection",
    "Get-RoleReferenceCollectionFromGroup",
    "Get-RoleReferenceCollectionFromUser",
    "Get-User",
    "Get-UserByName",
    "Get-UserCollection",
    "Get-UserGroup",
    "Get-UserGroupByName",
    "Get-UserGroupCollection",
    "Get-UserGroupMembershipCollection",
    "Get-UserMembershipCollection",
    "New-Session",
    "New-User",
    "New-UserGroup",
    "Remove-RoleFromGroup",
    "Remove-RoleFromUser",
    "Remove-User",
    "Remove-UserFromGroup",
    "Remove-UserGroup",
    "Reset-UserPassword",
    "Update-User",
    "Update-UserGroup"
)
Register-ArgumentCompleter -CommandName $CommandsWithTenant -ParameterName Tenant -ScriptBlock $script:CompleteTenant

$CommandsWithRole = @(
    "Add-RoleToGroup",
    "Add-RoleToUser",
    "Remove-RoleFromGroup",
    "Remove-RoleFromUser"
)
Register-ArgumentCompleter -CommandName $CommandsWithRole -ParameterName Role -ScriptBlock $script:CompleteRole

$CommandsWithRoles = @(
    "New-ServiceUser"
)
Register-ArgumentCompleter -CommandName $CommandsWithRoles -ParameterName Roles -ScriptBlock $script:CompleteRole


