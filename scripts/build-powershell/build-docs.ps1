[cmdletbinding()]
Param(
    [string] $Module = "./tools/PSc8y/dist/PSc8y",

    [string] $OutputFolder = "./docs/_pwsh/cmdlets",

    [switch] $Recreate
)

if (!(Get-Module -Name "platyPS")) {
    Install-Module -Name platyPS -Scope CurrentUser -Force
}

Import-Module platyPS
Import-Module $Module -Force

$ModuleName = $Module -replace ".+/(\w+)$", '$1'

if ($Recreate) {
    rm -Rf $OutputFolder
}

Function Get-Category {
    Param(
        [string] $Name
    )
    switch -Regex ($Name) {
        "alarm" { "Alarms" }
        "asset" { "Assets" }
        "dataBroker" { "DataBroker" }
        "audit" { "AuditRecords" }
        "operation" { "Operations" }
        "device" { "Devices" }
        "agent" { "Agents" }
        "event" { "Events" }
        "measurement" { "Measurements" }
        "supportedSeries" { "Measurements" }
        "managedObject" { "ManagedObjects" }
        "binary" { "Binaries" }
        "retention" { "RetentionRules" }
        "(tenant|system)Option" { "TenantOption" }
        "tenant" { "Tenants" }
        "application" { "Applications" }
        "externalid" { "ExternalIdentity" }
        "microservice" { "Microservices" }
        "user" { "Users" }
        "session" { "Sessions" }
        "group" { "User Groups" }
        "-(client|custom)" { "Client Helpers" }
        default { "Misc." }
    }
}

Function Invoke-FixMarkdownFormatting {
    [cmdletbinding()]
    Param(
        [string] $File
    )

    $script:inSection = $false
    $OutputText = (Get-Content $File) | ForEach-Object {
        if ($script:inSection -and ($_ -match "^## ")) {
            $script:inSection = $false
        }
        if ($_ -match "## Syntax") {
            $script:inSection = $true
        }
        if ($script:inSection) {
            $line = $_
            if ($line -match "^ ") {
                # if line already starts with a space, then don't add another line break
                $line = ($line -replace "^ ([\-\[])", "`t`$1")
            }

            $line -replace " ([\-\[])", "`n`t`$1"
        } else {
            $_
        }
    }
    # Fix any markdown escaping 
    $OutputText = $OutputText -replace '\\([`\[\]])', "`$1"
    $OutputText | Out-File $File

}

if (!(Test-Path $OutputFolder)) {
    New-Item -Path $OutputFolder -ItemType Directory -Force

    # -ModulePagePath ""
    [array] $commands = Get-Command -Module $ModuleName
    foreach ($command in $commands) {
        # $category = $command.Name -ireplace "\w+\-([A-Z][a-z0-9]+).*", "`$1"
        New-MarkdownHelp -Command $command.Name -OutputFolder $OutputFolder -Metadata @{
            'layout' = 'powershell'
            'category' = Get-Category $command.Name | Select-Object -First 1
            'title' = $command.Name
        }

        # Fix/customize markdown formatting
        Invoke-FixMarkdownFormatting -File "$OutputFolder/$($command.Name).md"
    }
    
} else {
    Update-MarkdownHelp $OutputFolder
}
