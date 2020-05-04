Function Get-C8ySessionProperty {
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [string] $Name
    )

    switch ($Name) {
        "tenant" {
            $env:C8Y_TENANT
        }

        "host" {
            $env:C8Y_HOST
        }
    }
}
