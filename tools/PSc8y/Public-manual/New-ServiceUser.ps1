Function New-ServiceUser {
    <#
    .SYNOPSIS
    New service user (via a dummy microservice user)

    .DESCRIPTION
    Create a new microservice application used to provide a service user used for external automation tasks

    .EXAMPLE
    PS> New-ServiceUser -Name "automation01" -RequiredRoles @("ROLE_INVENTORY_READ") -Tenants t123456

    Create a new microservice called automation01 which has permissions to read the inventory, and subscribe the application to tenant t123456

    .LINK
    c8y microservices serviceusers create

    .LINK
    Get-ServiceUser
    #>
    [cmdletbinding(PositionalBinding = $true,
                   HelpUri = '')]
    [Alias()]
    [OutputType([object])]
    Param(
        # Name of the microservice. An id is also accepted however the name have been previously uploaded.
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [string]
        $Name,

        # Roles which should be assigned to the service user, i.e. ROLE_INVENTORY_READ
        [Parameter(Mandatory = $false)]
        [string[]]
        $Roles,

        # Tenants IDs where the application should be subscribed. Useful when using in a multi tenant scenario where the
        # application is created in the management tenant, and a service user can be created via subscribing to the application on each
        # sub tenant
        [Parameter(Mandatory = $false)]
        [string[]]
        $Tenants
    )
    DynamicParam {
        Get-ClientCommonParameters -Type "Create"
    }

    Begin {

        if ($env:C8Y_DISABLE_INHERITANCE -ne $true) {
            # Inherit preference variables
            Use-CallerPreference -Cmdlet $PSCmdlet -SessionState $ExecutionContext.SessionState
        }

        $c8yargs = New-ClientArgument -Parameters $PSBoundParameters -Command "microservices serviceusers create"
        $ClientOptions = Get-ClientOutputOption $PSBoundParameters
        $TypeOptions = @{
            Type            = "application/json"
            ItemType        = ""
            BoundParameters = $PSBoundParameters
        }
    }

    Process {

        if ($ClientOptions.ConvertToPS) {
            c8y microservices serviceusers create $c8yargs `
            | ConvertFrom-ClientOutput @TypeOptions
        }
        else {
            c8y microservices serviceusers create $c8yargs
        }
    }

    End {}
}
