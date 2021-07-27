# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-AuditRecord {
<#
.SYNOPSIS
Create audit record

.DESCRIPTION
Create a new audit record for a given action

.LINK
https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y/auditrecords_create

.EXAMPLE
PS> New-AuditRecord -Type "ManagedObject" -Time "0s" -Text "Managed Object updated: my_Prop: value" -Source $Device.id -Activity "Managed Object updated" -Severity "information"

Create an audit record for a custom managed object update


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Identifies the type of this audit record.
        [Parameter()]
        [ValidateSet('Alarm','Application','BulkOperation','CepModule','Connector','Event','Group','Inventory','InventoryRole','Operation','Option','Report','SingleSignOn','SmartRule','SYSTEM','Tenant','TenantAuthConfig','TrustedCertificates','UserAuthentication')]
        [string]
        $Type,

        # Time of the audit record. Defaults to current timestamp.
        [Parameter()]
        [string]
        $Time,

        # Text description of the audit record.
        [Parameter()]
        [string]
        $Text,

        # An optional ManagedObject that the audit record originated from
        [Parameter(ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Source,

        # The activity that was carried out.
        [Parameter()]
        [string]
        $Activity,

        # The severity of action: critical, major, minor, warning or information.
        [Parameter()]
        [ValidateSet('critical','major','minor','warning','information')]
        [string]
        $Severity,

        # The user responsible for the audited action.
        [Parameter()]
        [string]
        $User,

        # The application used to carry out the audited action.
        [Parameter()]
        [string]
        $Application
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create", "Template"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "auditrecords create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type = "application/vnd.com.nsn.cumulocity.auditRecord+json"
            ItemType = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            $Source `
            | Group-ClientRequests `
            | c8y auditrecords create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Source `
            | Group-ClientRequests `
            | c8y auditrecords create $c8yargs
        }
        
    }

    End {}
}
