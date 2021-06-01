[cmdletbinding()]
Param()

# Convert the yaml specs to json
if (!(Get-Command yaml2json -ErrorAction SilentlyContinue)) {
    Write-Warning "Missing yamljs. Trying to install yamljs now"

    npm install -g yamljs
}

Write-Host "Converting yaml specs to json" -ForegroundColor Gray
$WorkDir = Resolve-Path -Path "$PSScriptRoot/../api/spec/yaml" -Relative
Get-ChildItem -Path "$WorkDir/*" -Include "*.yaml", "*.yml" | ForEach-Object {
    Write-Verbose ("Converting yaml spec {0}" -f $_.FullName)
    yaml2json -s -r -p $_.FullName
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Could not convert yaml spec to json. $LASTEXITCODE, file=$_"
    }
}

# Copy files to another folder
$DestSpecDirectory = Resolve-Path "$PSScriptRoot/../api/spec/json"
Move-Item -Path "$WorkDir/*.json" -Destination "$DestSpecDirectory/" -Force
