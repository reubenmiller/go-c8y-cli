# Code generated from specification version 1.0.0: DO NOT EDIT
Function New-AuditRecord {
<#
.SYNOPSIS
Create audit record

.DESCRIPTION
Create a new audit record for a given action

.LINK
c8y auditRecords create

.EXAMPLE
PS> New-AuditRecord -Type "ManagedObject" -Time "0s" -Text "Managed Object updated: my_Prop: value" -Source $Device.id -Activity "Managed Object updated" -Severity "information"

Create an audit record for a custom managed object update


#>
    [cmdletbinding(PositionalBinding=$true,
                   HelpUri='')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Identifies the type of this audit record. (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Type,

        # Time of the audit record. Defaults to current timestamp.
        [Parameter()]
        [string]
        $Time,

        # Text description of the audit record. (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Text,

        # An optional ManagedObject that the audit record originated from (required)
        [Parameter(Mandatory = $true,
                   ValueFromPipeline=$true,
                   ValueFromPipelineByPropertyName=$true)]
        [object[]]
        $Source,

        # The activity that was carried out. (required)
        [Parameter(Mandatory = $true)]
        [string]
        $Activity,

        # The severity of action: critical, major, minor, warning or information. (required)
        [Parameter(Mandatory = $true)]
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

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "auditRecords create"
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
            | c8y auditRecords create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            $Source `
            | Group-ClientRequests `
            | c8y auditRecords create $c8yargs
        }
        
    }

    End {}
}
