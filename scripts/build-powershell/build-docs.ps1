[cmdletbinding()]
Param(
    [string] $Module = "./tools/PSc8y/dist/PSc8y",

    [string] $OutputFolder = "",

    [string] $DocBaseUrl = "https://reubenmiller.github.io/go-c8y-cli/docs/cli/c8y",

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
        "-binary" { "Binaries" }
        "retention" { "RetentionRules" }
        "(tenant|system)Option" { "TenantOption" }
        "tenant" { "Tenants" }
        "tenantname" { "Tenants" }
        "application" { "Applications" }
        "externalid" { "ExternalIdentity" }
        "microservice" { "Microservices" }
        "-(test|current)?user" { "Users" }
        "role" { "Role" }
        "session" { "Sessions" }
        "group" { "User Groups" }
        "-(client|custom)" { "Client Helpers" }
        "configuration" { "Configuration" }
        "firmwarePatch" { "FirmwarePatch" }
        "firmwareVersion" { "FirmwareVersion" }
        "firmware" { "Firmware" }
        "softwareVersion" { "SoftwareVersion" }
        "software" { "Software" }
        "deviceprofiles" { "DeviceProfiles" }
        default { "Misc." }
    }
}

Function Invoke-FixMarkdownFormatting {
    [cmdletbinding()]
    Param(
        [string] $File,

        [string] $Name,

        [string] $BaseUrl
    )

    $script:inSection = $false
    $OutputText = (Get-Content $File) | ForEach-Object {
        $line = $_
        if ($script:inSection -and ($line -match "^## ")) {
            $script:inSection = $false
        }

        # Fix links with the full url
        # [c8y applications createHostedApplication]() => [baseurl/c8y/applications_createhostedapplication]
        if ($line -match "^\[c8y (.+)\]\(\)") {
            $pageName = ($Matches[1] -replace " ", "_").ToLower()
            $line = "[$BaseUrl/$pageName]($BaseUrl/$pageName)" 
        }

        if ($line -match "## Syntax") {
            $script:inSection = $true
        }
        if ($script:inSection) {
            if ($line -match "^ ") {
                # if line already starts with a space, then don't add another line break
                $line = ($line -replace "^ ([\-\[])", "`t`$1")
            }

            $line -replace " ([\-\[])", "`n`t`$1"
        } else {
            $line
        }
    }
    # Fix any markdown escaping 
    $OutputText = $OutputText -replace '\\([`\[\]])', "`$1"
    $OutputText = $OutputText -replace "##? $Name", ""
    $OutputText | Out-File $File

}

if (!(Test-Path $OutputFolder)) {
    New-Item -Path $OutputFolder -ItemType Directory -Force

    # -ModulePagePath ""
    # | Where-Object { $_.Name -eq 'New-HostedApplication' }
    [array] $commands = Get-Command -Module $ModuleName
    foreach ($command in $commands) {
        $category = Get-Category $command.Name | Select-Object -First 1
        $CurrentOutputFolder = $OutputFolder

        # Move to target folder
        if ($category) {
            $CurrentOutputFolder = "$OutputFolder/$category"
        }
        if ($CurrentOutputFolder -and (-Not (Test-Path $CurrentOutputFolder))) {
            $null = New-Item -Path "$CurrentOutputFolder" -ItemType Directory -Force
        }

        New-MarkdownHelp -Command $command.Name -OutputFolder $CurrentOutputFolder -Metadata @{
            # 'layout' = 'powershell'
            'category' = $category
            'id' = $command.Name
            'title' = $command.Name
            'slug' = "/docs/cli/psc8y/{0}/{1}" -f $category, $command.Name.ToLower()
        }

        # Fix/customize markdown formatting
        Invoke-FixMarkdownFormatting `
            -File "$CurrentOutputFolder/$($command.Name).md" `
            -Name $command.Name `
            -BaseUrl $DocBaseUrl
    }
    
} else {
    Update-MarkdownHelp $OutputFolder
}
