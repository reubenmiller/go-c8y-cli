Function New-TestAgent {
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $false,
            Position = 0
        )]
        [string] $Name = "testagent",

        [switch] $Force
    )
    $Data = @{
        c8y_IsDevice = @{}
        com_cumulocity_model_Agent = @{}
    }

    $AgentName = New-RandomString -Prefix "${Name}_"
    $TestAgent = PSc8y\New-ManagedObject `
        -Name $AgentName `
        -Data $Data `
        -Force:$Force

    $TestAgent
}
