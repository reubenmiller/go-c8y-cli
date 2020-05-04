Function Format-CommandParameter {
    [cmdletbinding()]
    Param(
        [Parameter(
            Mandatory = $true,
            Position = 0
        )]
        [object] $ParameterList
    )
    $Parameters = @{}

    $ParameterList = (Get-Command -Name $CommandName).Parameters

    # Grab each parameter value, using Get-Variable
    foreach ($Name in ($ParameterList.Keys -notmatch "^Raw$")) {
        $iParam = Get-Variable -Name $Name -ErrorAction SilentlyContinue;

        if ($iParam.Value -is [Switch]) {
            if ($iParam.Value.IsPresent -and $iParam) {
                $Parameters[$Name] = $true
            }
        } elseif ($iParam.Value -is [hashtable]) {
            $Parameters[$Name] = "{0}" -f ((ConvertTo-Json $iParam.Value -Compress) -replace '"', '\"')
        } elseif ($iParam.Value -is [datetime]) {
            $Parameters[$Name] = Format-Date $iParam.Value
        } else {
            if ("$iParam" -notmatch "^$") {
                $Parameters[$Name] = $iParam.Value
            }
        }
    }

    $Parameters
}
