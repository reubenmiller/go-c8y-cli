Function Get-ClientBinaryVersion {
    [cmdletbinding()]
    Param()
    $c8y = Get-ClientBinary
    & $c8y version
}
