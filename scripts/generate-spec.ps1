[cmdletbinding()]
Param()

# Convert the yaml specs to json
if (!(Get-Command yaml2json -ErrorAction SilentlyContinue)) {
    Write-Error "Missing yaml2json. Please install it using 'npm install -g yamljs'"
    return
}

Write-Host "Converting yaml specs to json" -ForegroundColor Gray
$WorkDir = Resolve-Path -Path "$PSScriptRoot/../api/spec/yaml" -Relative
yaml2json -s -r -p $WorkDir

# Copy files to another folder
$DestSpecDirectory = Resolve-Path "$PSScriptRoot/../api/spec/json"
Move-Item -Path "$WorkDir/*.json" -Destination "$DestSpecDirectory/" -Force
