#!/bin/bash

set -euo

arch=$(dpkg --print-architecture)

case $arch in
    arm64)
        arch="arm64"
        ;;

    *)
        arch=x64
        ;;

esac

PWSH_VERSION="7.4.6"

# pre-requisites
apt-get update && apt-get install -y libicu-dev

# Download the powershell '.tar.gz' archive
curl -L -o /tmp/powershell.tar.gz "https://github.com/PowerShell/PowerShell/releases/download/v${PWSH_VERSION}/powershell-${PWSH_VERSION}-linux-${arch}.tar.gz"

# Create the target folder where powershell will be placed
mkdir -p /opt/microsoft/powershell/7

# Expand powershell to the target folder
tar zxf /tmp/powershell.tar.gz -C /opt/microsoft/powershell/7

# Set execute permissions
chmod +x /opt/microsoft/powershell/7/pwsh

# Create the symbolic link that points to pwsh
ln -s /opt/microsoft/powershell/7/pwsh /usr/bin/pwsh
